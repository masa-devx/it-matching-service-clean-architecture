package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// 掲載状態。DBのCHECK制約（projects.status）と対応させる
const (
	projectStatusDraft     = "draft"
	projectStatusPublished = "published"
	projectStatusClosed    = "closed"
)

// projectStatuses は取りうる掲載状態の一覧。絞り込みクエリの検証に使う
var projectStatuses = []string{
	projectStatusDraft,
	projectStatusPublished,
	projectStatusClosed,
}

// isProjectStatus は文字列が掲載状態として定義済みかを返す
func isProjectStatus(s string) bool {
	return slices.Contains(projectStatuses, s)
}

type projectRequest struct {
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	RequiredSkills []string `json:"required_skills"`
	HourlyRateMin  int      `json:"hourly_rate_min"`
	HourlyRateMax  int      `json:"hourly_rate_max"`
	HoursPerWeek   int      `json:"hours_per_week"`
	RemoteOK       bool     `json:"remote_ok"`
	Status         string   `json:"status"`
}

type projectResponse struct {
	ID             int64     `json:"id"`
	CompanyID      int64     `json:"company_id"`
	CompanyName    string    `json:"company_name"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	RequiredSkills []string  `json:"required_skills"`
	HourlyRateMin  int       `json:"hourly_rate_min"`
	HourlyRateMax  int       `json:"hourly_rate_max"`
	HoursPerWeek   int       `json:"hours_per_week"`
	RemoteOK       bool      `json:"remote_ok"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

// handleCreateProject は POST /projects。企業ユーザーのみ（requireRole で担保）。
// company_id はリクエスト値ではなく、検証済みトークンの userID から引く（IDOR対策）
func handleCreateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", nil)
		return
	}

	var req projectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "リクエストボディが不正です", err)
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.RequiredSkills = normalizeSkills(req.RequiredSkills)
	if req.Status == "" {
		req.Status = projectStatusDraft
	}
	if msg := validateProject(req); msg != "" {
		writeError(r.Context(), w, http.StatusBadRequest, msg, nil)
		return
	}

	companyID, name, err := fetchCompany(userID)
	if errors.Is(err, sql.ErrNoRows) {
		// 掲載には会社情報が必要。原因と次の行動が分かる文言にする
		writeError(r.Context(), w, http.StatusBadRequest, "先に企業プロフィールを登録してください", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "案件の作成に失敗しました", err)
		return
	}

	res := projectResponse{CompanyID: companyID, CompanyName: name}
	err = db.QueryRow(
		`INSERT INTO projects
		   (company_id, title, description, required_skills, hourly_rate_min, hourly_rate_max, hours_per_week, remote_ok, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, title, description, required_skills, hourly_rate_min, hourly_rate_max, hours_per_week, remote_ok, status, created_at`,
		companyID, req.Title, req.Description, pgtype.FlatArray[string](req.RequiredSkills),
		req.HourlyRateMin, req.HourlyRateMax, req.HoursPerWeek, req.RemoteOK, req.Status,
	).Scan(
		&res.ID, &res.Title, &res.Description, pgtype.NewMap().SQLScanner(&res.RequiredSkills),
		&res.HourlyRateMin, &res.HourlyRateMax, &res.HoursPerWeek, &res.RemoteOK, &res.Status, &res.CreatedAt,
	)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "案件の作成に失敗しました", err)
		return
	}
	if res.RequiredSkills == nil {
		res.RequiredSkills = []string{}
	}

	writeJSON(w, http.StatusCreated, res)
}

// 一覧の取得件数。上限を設けないと「limit=100000」で全件を引かれる
const (
	defaultLimit = 20
	maxLimit     = 100
)

// projectListResponse は一覧のレスポンス。配列を直接返さず包むのは、
// 後からページング情報（次ページの有無・総件数）を足せるようにするため
type projectListResponse struct {
	Projects []projectResponse `json:"projects"`
	Total    int               `json:"total"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
}

// 一覧・詳細で共通の SELECT。会社名を JOIN で同時に取り、N+1（案件ごとに会社を引く）を避ける
const projectSelectSQL = `
	SELECT p.id, p.company_id, c.name,
	       p.title, p.description, p.required_skills,
	       p.hourly_rate_min, p.hourly_rate_max, p.hours_per_week,
	       p.remote_ok, p.status, p.created_at
	FROM projects p
	JOIN companies c ON c.id = p.company_id`

// projectFilter は一覧の検索条件。ゼロ値は「指定なし＝絞り込まない」を意味する
type projectFilter struct {
	Skills     []string // すべて含む案件に絞る（AND条件）
	RateMin    int      // 希望する時給の下限（案件の上限がこれ以上）
	RateMax    int      // 希望する時給の上限（案件の下限がこれ以下）
	HoursMax   int      // 週の稼働時間の上限
	RemoteOnly bool     // true のときだけリモート可に絞る
	Keyword    string   // タイトルの部分一致
}

const (
	maxFilterSkills  = 10
	maxKeywordLength = 100
	maxRate          = 1000000
	maxHoursPerWeek  = 168
)

// parseProjectFilter はクエリ文字列から検索条件を組み立てる。
// 数値が不正な場合は 0（指定なし）として扱い、範囲の妥当性は validateFilter で検証する
func parseProjectFilter(r *http.Request) projectFilter {
	q := r.URL.Query()
	return projectFilter{
		Skills:     normalizeSkills(strings.Split(q.Get("skills"), ",")),
		RateMin:    intQuery(r, "rate_min", 0, 0, math.MaxInt32),
		RateMax:    intQuery(r, "rate_max", 0, 0, math.MaxInt32),
		HoursMax:   intQuery(r, "hours_max", 0, 0, math.MaxInt32),
		RemoteOnly: q.Get("remote_ok") == "true",
		Keyword:    strings.TrimSpace(q.Get("q")),
	}
}

// validateFilter は検索条件の妥当性を検証する。
// 「無視して勝手に丸める」のではなく 400 で返し、意図と違う結果を黙って見せない
func validateFilter(f projectFilter) string {
	if len(f.Skills) > maxFilterSkills {
		return "スキルは10個以内で指定してください"
	}
	if len([]rune(f.Keyword)) > maxKeywordLength {
		return "キーワードは100文字以内で指定してください"
	}
	if f.RateMin > maxRate || f.RateMax > maxRate {
		return "単価は0〜1000000の範囲で指定してください"
	}
	// 両方指定されている場合のみ大小関係を検証する（片方だけの指定は有効）
	if f.RateMin > 0 && f.RateMax > 0 && f.RateMin > f.RateMax {
		return "単価の下限は上限以下で指定してください"
	}
	if f.HoursMax > maxHoursPerWeek {
		return "週の稼働時間は0〜168の範囲で指定してください"
	}
	return ""
}

// buildProjectWhere は検索条件から WHERE 句とプレースホルダの引数列を組み立てる。
// 値は必ず args に積み、SQL文字列にはプレースホルダ番号だけを書く
// （文字列連結で値を埋め込むと SQL インジェクションの穴になる）
func buildProjectWhere(f projectFilter) (string, []any) {
	// 一覧は常に公開中のものだけを対象にする
	args := []any{projectStatusPublished}
	conds := []string{"p.status = $1"}

	// 配列の包含演算子。GINインデックス（projects_required_skills_idx）が使われる
	if len(f.Skills) > 0 {
		args = append(args, pgtype.FlatArray[string](f.Skills))
		conds = append(conds, fmt.Sprintf("p.required_skills @> $%d", len(args)))
	}
	// 単価は「案件のレンジ」と「希望のレンジ」が重なるかで判定する。
	// 希望下限より案件の上限が高く、希望上限より案件の下限が低ければ交差している
	if f.RateMin > 0 {
		args = append(args, f.RateMin)
		conds = append(conds, fmt.Sprintf("p.hourly_rate_max >= $%d", len(args)))
	}
	if f.RateMax > 0 {
		args = append(args, f.RateMax)
		conds = append(conds, fmt.Sprintf("p.hourly_rate_min <= $%d", len(args)))
	}
	if f.HoursMax > 0 {
		args = append(args, f.HoursMax)
		conds = append(conds, fmt.Sprintf("p.hours_per_week <= $%d", len(args)))
	}
	// false は「リモートでなくてもよい」の意味なので、true のときだけ条件を足す
	if f.RemoteOnly {
		conds = append(conds, "p.remote_ok = true")
	}
	if f.Keyword != "" {
		// ILIKE は大文字小文字を区別しない部分一致。ワイルドカードは値側に付ける
		args = append(args, "%"+f.Keyword+"%")
		conds = append(conds, fmt.Sprintf("p.title ILIKE $%d", len(args)))
	}

	return " WHERE " + strings.Join(conds, " AND "), args
}

// handleListProjects は GET /projects。公開中の案件を検索条件で絞り、新しい順に返す
func handleListProjects(w http.ResponseWriter, r *http.Request) {
	limit := intQuery(r, "limit", defaultLimit, 1, maxLimit)
	offset := intQuery(r, "offset", 0, 0, math.MaxInt32)

	filter := parseProjectFilter(r)
	if msg := validateFilter(filter); msg != "" {
		writeError(r.Context(), w, http.StatusBadRequest, msg, nil)
		return
	}
	where, args := buildProjectWhere(filter)

	// 総件数（ページャの表示に必要）。JOIN 不要な条件しか無いため projects 単体で数える
	var total int
	countSQL := `SELECT count(*) FROM projects p` + where
	if err := db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "案件の取得に失敗しました", err)
		return
	}

	// LIMIT/OFFSET は条件の後ろに続くので、プレースホルダ番号は args の続きになる
	pageArgs := append(append([]any{}, args...), limit, offset)
	rows, err := db.Query(
		projectSelectSQL+where+fmt.Sprintf(`
		 ORDER BY p.created_at DESC, p.id DESC
		 LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2),
		pageArgs...,
	)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "案件の取得に失敗しました", err)
		return
	}
	defer func() { _ = rows.Close() }()

	projects := make([]projectResponse, 0, limit)
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "案件の取得に失敗しました", err)
			return
		}
		projects = append(projects, p)
	}
	// ループ中に発生したエラー（通信断など）は Err() で最後に確認する
	if err := rows.Err(); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "案件の取得に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusOK, projectListResponse{
		Projects: projects,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	})
}

// handleGetProject は GET /projects/{id}。公開中のもののみ返す（下書き・終了は404）
func handleGetProject(w http.ResponseWriter, r *http.Request) {
	// Go 1.22+ のパスパラメータ。ルート定義 "GET /projects/{id}" の {id} を取り出す
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(r.Context(), w, http.StatusNotFound, "案件が見つかりません", nil)
		return
	}

	row := db.QueryRow(projectSelectSQL+` WHERE p.id = $1 AND p.status = $2`, id, projectStatusPublished)
	p, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		// 未公開と存在しないを区別せず404にする（下書きの存在を外部に漏らさない）
		writeError(r.Context(), w, http.StatusNotFound, "案件が見つかりません", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "案件の取得に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusOK, p)
}

// scanner は *sql.Row と *sql.Rows の共通部分。1行のスキャン処理を共有するための最小インターフェース
type scanner interface {
	Scan(dest ...any) error
}

func scanProject(s scanner) (projectResponse, error) {
	var p projectResponse
	err := s.Scan(
		&p.ID, &p.CompanyID, &p.CompanyName,
		&p.Title, &p.Description, pgtype.NewMap().SQLScanner(&p.RequiredSkills),
		&p.HourlyRateMin, &p.HourlyRateMax, &p.HoursPerWeek,
		&p.RemoteOK, &p.Status, &p.CreatedAt,
	)
	if err != nil {
		return p, err
	}
	if p.RequiredSkills == nil {
		p.RequiredSkills = []string{}
	}
	return p, nil
}

// intQuery はクエリ文字列の整数を範囲内に丸めて返す。不正値は既定値にフォールバックする
func intQuery(r *http.Request, key string, fallback, minValue, maxValue int) int {
	v, err := strconv.Atoi(r.URL.Query().Get(key))
	if err != nil {
		return fallback
	}
	return min(max(v, minValue), maxValue)
}

// fetchCompany は userID に紐づく企業プロフィールを返す（未作成なら sql.ErrNoRows）
func fetchCompany(userID int64) (int64, string, error) {
	var (
		id   int64
		name string
	)
	err := db.QueryRow(
		`SELECT id, name FROM companies WHERE user_id = $1`,
		userID,
	).Scan(&id, &name)
	return id, name, err
}

// validateProject は案件の入力検証
func validateProject(req projectRequest) string {
	if req.Title == "" {
		return "案件タイトルは必須です"
	}
	if len([]rune(req.Title)) > 100 {
		return "案件タイトルは100文字以内にしてください"
	}
	if len([]rune(req.Description)) > 5000 {
		return "案件内容は5000文字以内にしてください"
	}
	if len(req.RequiredSkills) > 30 {
		return "必須スキルは30個以内にしてください"
	}
	for _, s := range req.RequiredSkills {
		if len([]rune(s)) > 50 {
			return "各スキルは50文字以内にしてください"
		}
	}
	if req.HourlyRateMin < 0 || req.HourlyRateMax < 0 ||
		req.HourlyRateMin > 1000000 || req.HourlyRateMax > 1000000 {
		return "単価は0〜1000000の範囲で入力してください"
	}
	// DBのCHECK制約と同じルールをアプリ側でも検証し、意味のあるメッセージを返す
	if req.HourlyRateMin > req.HourlyRateMax {
		return "単価の下限は上限以下にしてください"
	}
	if req.HoursPerWeek < 0 || req.HoursPerWeek > 168 {
		return "週の稼働時間は0〜168の範囲で入力してください"
	}
	if req.Status != projectStatusDraft && req.Status != projectStatusPublished &&
		req.Status != projectStatusClosed {
		return "status は draft / published / closed のいずれかを指定してください"
	}
	return ""
}
