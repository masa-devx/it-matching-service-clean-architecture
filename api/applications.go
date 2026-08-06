package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// pgUniqueViolation は PostgreSQL の一意制約違反。
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const pgUniqueViolation = "23505"

type applicationRequest struct {
	Message string `json:"message"`
}

type applicationResponse struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	TalentID  int64     `json:"talent_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// handleCreateApplication は POST /projects/{id}/applications。人材ユーザーのみ（requireRole で担保）。
// 企業ロールはミドルウェアで弾かれるため、自社案件への応募は構造的に発生しない。
// talent_id はリクエスト値ではなく、検証済みトークンの userID から引く（IDOR対策）
func handleCreateApplication(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", nil)
		return
	}

	projectID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(r.Context(), w, http.StatusNotFound, "案件が見つかりません", nil)
		return
	}

	var req applicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "リクエストボディが不正です", err)
		return
	}

	req.Message = strings.TrimSpace(req.Message)
	if msg := validateApplication(req); msg != "" {
		writeError(r.Context(), w, http.StatusBadRequest, msg, nil)
		return
	}

	talentID, err := fetchTalent(userID)
	if errors.Is(err, sql.ErrNoRows) {
		// 応募には人材プロフィール（名前・スキル）が必要。次の行動が分かる文言にする
		writeError(r.Context(), w, http.StatusBadRequest, "先に人材プロフィールを登録してください", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "応募に失敗しました", err)
		return
	}

	// 公開中の案件だけを応募対象にする。下書き・募集終了は「存在しない」扱い
	// （handleGetProject と同じ方針。未公開案件の存在を外部に漏らさない）
	var exists bool
	err = db.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM projects WHERE id = $1 AND status = $2)`,
		projectID, projectStatusPublished,
	).Scan(&exists)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "応募に失敗しました", err)
		return
	}
	if !exists {
		writeError(r.Context(), w, http.StatusNotFound, "案件が見つかりません", nil)
		return
	}

	res := applicationResponse{ProjectID: projectID, TalentID: talentID}
	err = db.QueryRow(
		`INSERT INTO applications (project_id, talent_id, message)
		 VALUES ($1, $2, $3)
		 RETURNING id, status, message, created_at`,
		projectID, talentID, req.Message,
	).Scan(&res.ID, &res.Status, &res.Message, &res.CreatedAt)

	// 重複は事前SELECTではなくUNIQUE違反で検知する。
	// 「確認してから挿入」は確認と挿入の間に別リクエストが入ると破れるため（TOCTOU）
	if isUniqueViolation(err) {
		writeError(r.Context(), w, http.StatusConflict, "この案件にはすでに応募しています", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "応募に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusCreated, res)
}

// talentApplicationResponse は人材から見た応募（「どの案件に応募したか」が主役）。
// 同じ applications テーブルでも、企業から見た形（誰が応募してきたか）とは別の型にする
type talentApplicationResponse struct {
	ID           int64     `json:"id"`
	ProjectID    int64     `json:"project_id"`
	ProjectTitle string    `json:"project_title"`
	CompanyName  string    `json:"company_name"`
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at"`
}

type talentApplicationListResponse struct {
	Applications []talentApplicationResponse `json:"applications"`
	Total        int                         `json:"total"`
	Limit        int                         `json:"limit"`
	Offset       int                         `json:"offset"`
}

// 表示に必要な案件名・企業名を一度の結合で取る（件数分のクエリを発行しない＝N+1回避）。
// talents を経由する結合が所有チェックも兼ねており、
// トークンの持ち主以外の応募は数えることも読むこともできない。
//
// $2 は選考状態の絞り込み。空文字なら条件が無効になるため、
// SQL文字列を条件によって組み替える必要がない（＝連結の余地を作らない）
const myApplicationsFrom = `
	FROM applications a
	JOIN talents t   ON t.id = a.talent_id
	JOIN projects p  ON p.id = a.project_id
	JOIN companies c ON c.id = p.company_id
	WHERE t.user_id = $1
	  AND ($2 = '' OR a.status = $2)`

const myApplicationsOrder = `
	ORDER BY a.created_at DESC, a.id DESC
	LIMIT $3 OFFSET $4`

// handleListMyApplications は GET /me/applications。人材の応募履歴。
// 案件名・企業名はループ内で引かずJOINで一度に取る（N+1を作らない）
func handleListMyApplications(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", nil)
		return
	}

	status := r.URL.Query().Get("status")
	if status != "" && !isApplicationStatus(status) {
		writeError(r.Context(), w, http.StatusBadRequest, "選考状態の指定が不正です", nil)
		return
	}
	limit := intQuery(r, "limit", defaultLimit, 1, maxLimit)
	offset := intQuery(r, "offset", 0, 0, math.MaxInt32)

	// SQLは定数だけで組み立て、可変部分はすべてプレースホルダに渡す。
	// 空文字を「絞り込みなし」として扱えば、status の値がSQL文字列に混ざる経路が無くなる
	// （$2 IS NULL のような条件分岐をSQL側に持たせる書き方）
	args := []any{userID, status}

	var total int
	if err := db.QueryRow(`SELECT count(*)`+myApplicationsFrom, args...).Scan(&total); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "応募履歴の取得に失敗しました", err)
		return
	}

	rows, err := db.Query(
		`SELECT a.id, a.project_id, p.title, c.name, a.status, a.message, a.created_at`+
			myApplicationsFrom+myApplicationsOrder,
		userID, status, limit, offset,
	)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "応募履歴の取得に失敗しました", err)
		return
	}
	defer func() { _ = rows.Close() }()

	list := make([]talentApplicationResponse, 0, limit)
	for rows.Next() {
		var a talentApplicationResponse
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.ProjectTitle, &a.CompanyName,
			&a.Status, &a.Message, &a.CreatedAt); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "応募履歴の取得に失敗しました", err)
			return
		}
		list = append(list, a)
	}
	if err := rows.Err(); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "応募履歴の取得に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusOK, talentApplicationListResponse{
		Applications: list, Total: total, Limit: limit, Offset: offset,
	})
}

type applicationStatusRequest struct {
	Status string `json:"status"`
}

// handleUpdateApplicationStatus は PATCH /applications/{id}/status。
// 企業・人材のどちらも叩くため requireRole は付けず、実行してよいかは遷移表が判定する。
//
// 認可は3層で確認する: 認証（requireAuth）→ 所有（取得クエリのJOINに自分のIDを入れる）
// → 遷移可否（canTransition）。どれか1つでも欠けると他人の選考を進められてしまう
func handleUpdateApplicationStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", nil)
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(r.Context(), w, http.StatusNotFound, "応募が見つかりません", nil)
		return
	}

	var req applicationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "リクエストボディが不正です", err)
		return
	}

	role, err := fetchRole(userID)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "選考状態の更新に失敗しました", err)
		return
	}

	current, err := fetchApplicationStatus(id, userID, role)
	if errors.Is(err, sql.ErrNoRows) {
		// 他人の応募は「存在しない」として扱う（他社の選考状況を漏らさない）
		writeError(r.Context(), w, http.StatusNotFound, "応募が見つかりません", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "選考状態の更新に失敗しました", err)
		return
	}

	if !canTransition(current, req.Status, role) {
		// 409: 入力も権限も正しいが、現在の状態からは実行できない
		writeError(r.Context(), w, http.StatusConflict, "この応募に対してその操作はできません", nil)
		return
	}

	res, err := updateApplicationStatus(id, current, req.Status, role)
	if errors.Is(err, sql.ErrNoRows) {
		// 状態を読んでから更新するまでの間に他者が変更した（楽観ロックの検知）
		writeError(r.Context(), w, http.StatusConflict, "他の操作により状態が変わりました。画面を更新してください", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "選考状態の更新に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// fetchApplicationStatus は自分が関与する応募の現在の状態を返す。
// 所有チェックを WHERE に埋め込むことで、他人の応募はそもそも1行も返らない
// （取得してから所有を判定する形にすると、判定漏れがそのまま権限昇格になる）
func fetchApplicationStatus(id, userID int64, role string) (string, error) {
	var query string
	switch role {
	case roleCompany:
		query = `SELECT a.status FROM applications a
		         JOIN projects p ON p.id = a.project_id
		         JOIN companies c ON c.id = p.company_id
		         WHERE a.id = $1 AND c.user_id = $2`
	case roleTalent:
		query = `SELECT a.status FROM applications a
		         JOIN talents t ON t.id = a.talent_id
		         WHERE a.id = $1 AND t.user_id = $2`
	default:
		return "", sql.ErrNoRows
	}

	var status string
	err := db.QueryRow(query, id, userID).Scan(&status)
	return status, err
}

// updateApplicationStatus は状態を進める。
// WHERE に読み取り時の状態(from)を含めることで、読んでから更新するまでの間に
// 他者が変更していた場合は0行更新となり sql.ErrNoRows で検知できる（楽観ロック）
func updateApplicationStatus(id int64, from, to, role string) (applicationResponse, error) {
	var res applicationResponse
	// 意思表示の時刻は実行者側の列にだけ記録する。列名はプレースホルダにできないため
	// CASE で選ぶ（role を文字列連結するとSQLインジェクションの穴になる）
	err := db.QueryRow(
		`UPDATE applications
		 SET status = $1,
		     updated_at = now(),
		     company_acted_at = CASE WHEN $4 = 'company' THEN now() ELSE company_acted_at END,
		     talent_acted_at  = CASE WHEN $4 = 'talent'  THEN now() ELSE talent_acted_at  END
		 WHERE id = $2 AND status = $3
		 RETURNING id, project_id, talent_id, status, message, created_at`,
		to, id, from, role,
	).Scan(&res.ID, &res.ProjectID, &res.TalentID, &res.Status, &res.Message, &res.CreatedAt)
	return res, err
}

// isUniqueViolation は一意制約違反かどうかを判定する。
// ドライバ固有のエラー型を取り出すため errors.As を使う（文字列一致は環境差で壊れる）
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}

// fetchTalent は userID から人材プロフィールのIDを引く。
// 未登録なら sql.ErrNoRows（呼び出し側で「先に登録してください」に変換する）
func fetchTalent(userID int64) (int64, error) {
	var id int64
	err := db.QueryRow(`SELECT id FROM talents WHERE user_id = $1`, userID).Scan(&id)
	return id, err
}

// validateApplication は応募の入力検証
func validateApplication(req applicationRequest) string {
	if req.Message == "" {
		return "志望動機は必須です"
	}
	if len([]rune(req.Message)) > 2000 {
		return "志望動機は2000文字以内にしてください"
	}
	return ""
}
