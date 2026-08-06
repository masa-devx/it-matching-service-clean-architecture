package main

import (
	"database/sql"
	"encoding/json"
	"errors"
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
