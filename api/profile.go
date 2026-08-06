package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

type companyProfile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Industry    string `json:"industry"`
	Size        string `json:"size"`
}

type talentProfile struct {
	// 企業に見せる名前。companies.name と対になる（応募者一覧・応募履歴で使う）
	DisplayName           string   `json:"display_name"`
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
		writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", nil)
		return
	}

	role, err := fetchRole(userID)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "プロフィールの取得に失敗しました", err)
		return
	}

	profile, err := fetchProfile(userID, role)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "プロフィールの取得に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusOK, profileResponse{Role: role, Profile: profile})
}

// handlePutProfile は PUT /me/profile。作成と更新を1エンドポイントに統合する（upsert）。
// user_id は必ず検証済みトークンから取るため、他人のプロフィールは操作できない（IDOR対策）
func handlePutProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", nil)
		return
	}

	role, err := fetchRole(userID)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "プロフィールの保存に失敗しました", err)
		return
	}

	switch role {
	case roleCompany:
		var req companyProfile
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "リクエストボディが不正です", err)
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		if msg := validateCompanyProfile(req); msg != "" {
			writeError(r.Context(), w, http.StatusBadRequest, msg, nil)
			return
		}
		if err := upsertCompanyProfile(userID, req); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "プロフィールの保存に失敗しました", err)
			return
		}
		writeJSON(w, http.StatusOK, profileResponse{Role: role, Profile: req})

	case roleTalent:
		var req talentProfile
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(r.Context(), w, http.StatusBadRequest, "リクエストボディが不正です", err)
			return
		}
		req.DisplayName = strings.TrimSpace(req.DisplayName)
		req.Skills = normalizeSkills(req.Skills)
		if msg := validateTalentProfile(req); msg != "" {
			writeError(r.Context(), w, http.StatusBadRequest, msg, nil)
			return
		}
		if err := upsertTalentProfile(userID, req); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "プロフィールの保存に失敗しました", err)
			return
		}
		writeJSON(w, http.StatusOK, profileResponse{Role: role, Profile: req})

	default:
		writeError(r.Context(), w, http.StatusInternalServerError, "プロフィールの保存に失敗しました", nil)
	}
}

// upsertCompanyProfile は user_id の一意制約を利用した upsert（作成 or 更新）
func upsertCompanyProfile(userID int64, p companyProfile) error {
	_, err := db.Exec(
		`INSERT INTO companies (user_id, name, description, industry, size)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id) DO UPDATE
		 SET name        = EXCLUDED.name,
		     description = EXCLUDED.description,
		     industry    = EXCLUDED.industry,
		     size        = EXCLUDED.size,
		     updated_at  = now()`,
		userID, p.Name, p.Description, p.Industry, p.Size,
	)
	return err
}

func upsertTalentProfile(userID int64, p talentProfile) error {
	_, err := db.Exec(
		`INSERT INTO talents (user_id, display_name, bio, skills, years_of_exp, available_hours_per_week, desired_hourly_rate, remote_ok)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (user_id) DO UPDATE
		 SET display_name             = EXCLUDED.display_name,
		     bio                      = EXCLUDED.bio,
		     skills                   = EXCLUDED.skills,
		     years_of_exp             = EXCLUDED.years_of_exp,
		     available_hours_per_week = EXCLUDED.available_hours_per_week,
		     desired_hourly_rate      = EXCLUDED.desired_hourly_rate,
		     remote_ok                = EXCLUDED.remote_ok,
		     updated_at               = now()`,
		userID, p.DisplayName, p.Bio, pgtype.FlatArray[string](p.Skills), p.YearsOfExp,
		p.AvailableHoursPerWeek, p.DesiredHourlyRate, p.RemoteOK,
	)
	return err
}

// normalizeSkills は前後の空白除去・空要素の除去・重複排除を行う（入力の揺れを保存前に吸収する）
func normalizeSkills(skills []string) []string {
	seen := make(map[string]struct{}, len(skills))
	normalized := make([]string, 0, len(skills))
	for _, s := range skills {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		normalized = append(normalized, s)
	}
	return normalized
}

// validateCompanyProfile は企業プロフィールの入力検証（問題があればユーザー向け文言を返す）
func validateCompanyProfile(p companyProfile) string {
	if p.Name == "" {
		return "会社名は必須です"
	}
	if len([]rune(p.Name)) > 100 {
		return "会社名は100文字以内にしてください"
	}
	if len([]rune(p.Description)) > 2000 {
		return "会社説明は2000文字以内にしてください"
	}
	return ""
}

// validateTalentProfile は人材プロフィールの入力検証
func validateTalentProfile(p talentProfile) string {
	// DBは既存行のために空文字を許すが、保存時はここで必須にする（段階的な必須化）
	if p.DisplayName == "" {
		return "表示名は必須です"
	}
	if len([]rune(p.DisplayName)) > 50 {
		return "表示名は50文字以内にしてください"
	}
	if len([]rune(p.Bio)) > 2000 {
		return "自己紹介は2000文字以内にしてください"
	}
	if len(p.Skills) > 30 {
		return "スキルは30個以内にしてください"
	}
	for _, s := range p.Skills {
		if len([]rune(s)) > 50 {
			return "各スキルは50文字以内にしてください"
		}
	}
	if p.YearsOfExp < 0 || p.YearsOfExp > 70 {
		return "経験年数は0〜70の範囲で入力してください"
	}
	// 1週間は168時間。それを超える稼働は入力ミス
	if p.AvailableHoursPerWeek < 0 || p.AvailableHoursPerWeek > 168 {
		return "週の稼働可能時間は0〜168の範囲で入力してください"
	}
	if p.DesiredHourlyRate < 0 || p.DesiredHourlyRate > 1000000 {
		return "希望時給は0〜1000000の範囲で入力してください"
	}
	return ""
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
		// database/sql の Scan は基本型しか知らないため、TEXT[] は pgx の型マップに
		// sql.Scanner としてブリッジさせる（SQLScanner が []string への変換を担う）
		var skills []string
		err := db.QueryRow(
			`SELECT display_name, bio, skills, years_of_exp, available_hours_per_week, desired_hourly_rate, remote_ok
			 FROM talents WHERE user_id = $1`,
			userID,
		).Scan(&p.DisplayName, &p.Bio, pgtype.NewMap().SQLScanner(&skills), &p.YearsOfExp, &p.AvailableHoursPerWeek, &p.DesiredHourlyRate, &p.RemoteOK)
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
