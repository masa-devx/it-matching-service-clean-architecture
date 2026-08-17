package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
)

const (
	defaultProjectLimit int32 = 20
	maxProjectLimit     int32 = 50
)

// ErrProjectNotFound は不存在と未公開（draft / closed）を区別しない（未公開は取得しない原則・404）
var ErrProjectNotFound = errors.New("案件が見つかりません")

// Project は公開案件の閲覧（talent 視点）の業務ロジック
type Project struct {
	queries *db.Queries
}

func NewProject(queries *db.Queries) *Project {
	return &Project{queries: queries}
}

type ListPublishedParams struct {
	Cursor        *int64
	Limit         *int32
	Skills        []string
	RemoteOk      *bool
	MinHourlyRate *int32
}

type ProjectPage struct {
	Projects   []db.Project
	NextCursor *int64
}

// ListPublished は公開中の案件を seek 法で1ページ返す。
// limit+1 件取得し、余った1件の存在＝次ページあり・その id が次の読み出し起点（n+1 テクニック）
func (u *Project) ListPublished(ctx context.Context, p ListPublishedParams) (ProjectPage, error) {
	// limit は既定値と上限にクランプする（巨大な limit は DoS の入口・不正値は既定に落とす）
	limit := defaultProjectLimit
	if p.Limit != nil && *p.Limit > 0 {
		limit = min(*p.Limit, maxProjectLimit)
	}

	// 空のスキル指定は「絞り込みなし」に正規化（nil にすると narg が条件ごと無効化する）
	skills := p.Skills
	if len(skills) == 0 {
		skills = nil
	}

	rows, err := u.queries.ListPublishedProjects(ctx, db.ListPublishedProjectsParams{
		Cursor:        p.Cursor,
		Skills:        skills,
		RemoteOk:      p.RemoteOk,
		MinHourlyRate: p.MinHourlyRate,
		LimitPlusOne:  limit + 1,
	})
	if err != nil {
		return ProjectPage{}, fmt.Errorf("公開案件一覧の取得に失敗: %w", err)
	}

	page := ProjectPage{Projects: rows}
	if len(rows) > int(limit) {
		// n+1 件目＝次ページの先頭。その id をそのまま次の cursor にする（id <= cursor で続きが読める）
		next := rows[limit].ID
		page.Projects = rows[:limit]
		page.NextCursor = &next
	}
	return page, nil
}

// GetPublished は公開中の案件詳細を返す（未公開・不存在は同じ ErrProjectNotFound）
func (u *Project) GetPublished(ctx context.Context, projectID int64) (db.Project, error) {
	proj, err := u.queries.GetPublishedProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Project{}, ErrProjectNotFound
		}
		return db.Project{}, fmt.Errorf("案件取得に失敗: %w", err)
	}
	return proj, nil
}
