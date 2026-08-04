package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
)

type companyProfile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Industry    string `json:"industry"`
	Size        string `json:"size"`
}

type talentProfile struct {
	Bio                   string   `json:"bio"`
	Skills                []string `json:"skills"`
	YearsOfExp            int      `json:"years_of_exp"`
	AvailableHoursPerWeek int      `json:"available_hours_per_week"`
	DesiredHourlyRate     int      `json:"desired_hourly_rate"`
	RemoteOK              bool     `json:"remote_ok"`
}

// profileResponse は role と本体を包む。未作成のときは Profile が null になる
// （omitempty は付けない: キー自体を消すとフロントが「未作成」と「壊れた応答」を区別できない）
type profileResponse struct {
	Role    string `json:"role"`
	Profile any    `json:"profile"`
}

// handleGetProfile は GET /me/profile。role に応じて企業/人材のプロフィールを返す。
// 未作成は 404 ではなく 200 + null（「まだ無い」は正常な状態であり、エラーではない）
func handleGetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "認証が必要です", nil)
		return
	}

	role, err := fetchRole(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "プロフィールの取得に失敗しました", err)
		return
	}

	profile, err := fetchProfile(userID, role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "プロフィールの取得に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusOK, profileResponse{Role: role, Profile: profile})
}

// fetchRole は users からロールを取得する
func fetchRole(userID int64) (string, error) {
	var role string
	err := db.QueryRow(`SELECT role FROM users WHERE id = $1`, userID).Scan(&role)
	return role, err
}

// fetchProfile は role に対応するプロフィールを返す。未作成なら nil（＝JSONのnull）
func fetchProfile(userID int64, role string) (any, error) {
	switch role {
	case roleCompany:
		var p companyProfile
		err := db.QueryRow(
			`SELECT name, description, industry, size FROM companies WHERE user_id = $1`,
			userID,
		).Scan(&p.Name, &p.Description, &p.Industry, &p.Size)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return p, nil

	case roleTalent:
		var p talentProfile
		// TEXT[] は pgtype.FlatArray でスキャンする（database/sql は配列型を直接扱えない）
		var skills pgtype.FlatArray[string]
		err := db.QueryRow(
			`SELECT bio, skills, years_of_exp, available_hours_per_week, desired_hourly_rate, remote_ok
			 FROM talents WHERE user_id = $1`,
			userID,
		).Scan(&p.Bio, &skills, &p.YearsOfExp, &p.AvailableHoursPerWeek, &p.DesiredHourlyRate, &p.RemoteOK)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		// JSONで [] を返すため nil を空スライスに正規化する（null と [] の混在を避ける）
		p.Skills = skills
		if p.Skills == nil {
			p.Skills = []string{}
		}
		return p, nil
	}

	// CHECK制約があるため通常到達しない
	return nil, nil
}
