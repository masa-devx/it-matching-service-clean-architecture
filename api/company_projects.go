package main

import (
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// myProjectResponse は企業から見た自社案件。公開中だけを扱う一覧（projectResponse）と違い、
// 下書き・募集終了も含み、選考の進み具合（応募件数）を持つ
type myProjectResponse struct {
	ID             int64     `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	RequiredSkills []string  `json:"required_skills"`
	HourlyRateMin  int       `json:"hourly_rate_min"`
	HourlyRateMax  int       `json:"hourly_rate_max"`
	HoursPerWeek   int       `json:"hours_per_week"`
	RemoteOK       bool      `json:"remote_ok"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	// 応募の総数と、まだ企業が対応していない数（applied）。
	// 一覧で「どの案件に対応が残っているか」を判断するために返す
	ApplicationsCount int `json:"applications_count"`
	PendingCount      int `json:"pending_count"`
}

type myProjectListResponse struct {
	Projects []myProjectResponse `json:"projects"`
	Total    int                 `json:"total"`
	Limit    int                 `json:"limit"`
	Offset   int                 `json:"offset"`
}

// 自社案件の絞り込み。$1=user_id / $2=掲載状態（空文字は「絞り込まない」）。
// 条件によってSQL文字列を組み替えないことで、連結の余地を作らない
const myProjectsFrom = `
	FROM projects p
	JOIN companies c ON c.id = p.company_id
	WHERE c.user_id = $1
	  AND ($2 = '' OR p.status = $2)`

// 応募件数の集計。
//
// LEFT JOIN は必須：INNER にすると「応募が1件も無い案件」が一覧から消える
// （企業にとっては、まだ応募が来ていない案件こそ確認したい）。
//
// count(*) ではなく count(a.id) を使う理由も同じで、LEFT JOIN が作る NULL 行を
// 1件と数えてしまわないようにするため（count(列) は NULL を数えない）。
//
// FILTER 句を使うと、同じ1回のスキャンから「全件」と「未対応だけ」の2つの集計が取れる
const myProjectsSelectSQL = `
	SELECT p.id, p.title, p.description, p.required_skills,
	       p.hourly_rate_min, p.hourly_rate_max, p.hours_per_week,
	       p.remote_ok, p.status, p.created_at,
	       count(a.id)                                    AS applications_count,
	       count(a.id) FILTER (WHERE a.status = 'applied') AS pending_count
	FROM projects p
	JOIN companies c ON c.id = p.company_id
	LEFT JOIN applications a ON a.project_id = p.id
	WHERE c.user_id = $1
	  AND ($2 = '' OR p.status = $2)
	GROUP BY p.id, c.id
	ORDER BY p.created_at DESC, p.id DESC
	LIMIT $3 OFFSET $4`

// handleListMyProjects は GET /me/projects。企業ユーザーのみ（requireRole で担保）。
// company_id はリクエスト値ではなく、検証済みトークンの userID から辿る（IDOR対策）
func handleListMyProjects(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", nil)
		return
	}

	status := r.URL.Query().Get("status")
	if status != "" && !isProjectStatus(status) {
		writeError(r.Context(), w, http.StatusBadRequest, "掲載状態の指定が不正です", nil)
		return
	}
	limit := intQuery(r, "limit", defaultLimit, 1, maxLimit)
	offset := intQuery(r, "offset", 0, 0, math.MaxInt32)

	// 総件数は applications を結合せずに数える（1案件が応募数だけ重複しないように）
	var total int
	if err := db.QueryRow(`SELECT count(*)`+myProjectsFrom, userID, status).Scan(&total); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "案件の取得に失敗しました", err)
		return
	}

	rows, err := db.Query(myProjectsSelectSQL, userID, status, limit, offset)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "案件の取得に失敗しました", err)
		return
	}
	defer func() { _ = rows.Close() }()

	projects := make([]myProjectResponse, 0, limit)
	for rows.Next() {
		var p myProjectResponse
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, pgtype.NewMap().SQLScanner(&p.RequiredSkills),
			&p.HourlyRateMin, &p.HourlyRateMax, &p.HoursPerWeek,
			&p.RemoteOK, &p.Status, &p.CreatedAt,
			&p.ApplicationsCount, &p.PendingCount,
		); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "案件の取得に失敗しました", err)
			return
		}
		// JSONで [] を返すため nil を空スライスに正規化する（null と [] の混在を避ける）
		if p.RequiredSkills == nil {
			p.RequiredSkills = []string{}
		}
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "案件の取得に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusOK, myProjectListResponse{
		Projects: projects, Total: total, Limit: limit, Offset: offset,
	})
}

// 1件取得。一覧（myProjectsSelectSQL）と同じ集計を、案件IDで絞って行う。
// $2 が空文字のときに全状態を通す条件は不要なので、WHERE は所有チェックとIDだけ
const myProjectSelectSQL = `
	SELECT p.id, p.title, p.description, p.required_skills,
	       p.hourly_rate_min, p.hourly_rate_max, p.hours_per_week,
	       p.remote_ok, p.status, p.created_at,
	       count(a.id)                                    AS applications_count,
	       count(a.id) FILTER (WHERE a.status = 'applied') AS pending_count
	FROM projects p
	JOIN companies c ON c.id = p.company_id
	LEFT JOIN applications a ON a.project_id = p.id
	WHERE p.id = $1 AND c.user_id = $2
	GROUP BY p.id, c.id`

// handleGetMyProject は GET /me/projects/{id}。企業ユーザーのみ。
//
// 公開中しか返さない GET /projects/{id} とは別に用意する。
// 下書き・募集終了の案件も編集・状態変更の対象になるため、所有者には見せる必要がある
func handleGetMyProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", nil)
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(r.Context(), w, http.StatusNotFound, "案件が見つかりません", nil)
		return
	}

	var p myProjectResponse
	err = db.QueryRow(myProjectSelectSQL, id, userID).Scan(
		&p.ID, &p.Title, &p.Description, pgtype.NewMap().SQLScanner(&p.RequiredSkills),
		&p.HourlyRateMin, &p.HourlyRateMax, &p.HoursPerWeek,
		&p.RemoteOK, &p.Status, &p.CreatedAt,
		&p.ApplicationsCount, &p.PendingCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		// 他社の案件は「存在しない」として扱う（403だと存在が漏れる）
		writeError(r.Context(), w, http.StatusNotFound, "案件が見つかりません", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "案件の取得に失敗しました", err)
		return
	}
	if p.RequiredSkills == nil {
		p.RequiredSkills = []string{}
	}

	writeJSON(w, http.StatusOK, p)
}
