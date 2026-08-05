package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
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
