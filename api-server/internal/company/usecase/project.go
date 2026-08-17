// Package usecase は company API の業務ロジックを置く。
// 1つの公開メソッドが1つの業務操作を完結させる。複数テーブルへの書き込みが必要になったら、
// この層がトランザクション境界を持つ（Queries.WithTx で束ねる）。
package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
)

// Project は案件の業務ロジック。依存は生成された Queries のみ（HTTP の型を知らない）
type Project struct {
	queries *db.Queries
}

func NewProject(queries *db.Queries) *Project {
	return &Project{queries: queries}
}

// Create は案件を draft として作成する。
// 所有者（company_id）はクライアントから受け取らず、検証済みトークンの userID から
// プロフィールを引いて決める——「他社を所有者に指定する」形が存在しない（IDOR対策）
func (u *Project) Create(ctx context.Context, userID int64, params db.CreateProjectParams) (db.Project, error) {
	comp, err := u.queries.GetCompanyByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// トークンは有効だがプロフィールの実体が無い（削除済み等）
			return db.Project{}, ErrAuthFailed
		}
		return db.Project{}, fmt.Errorf("企業プロフィール取得に失敗: %w", err)
	}
	params.CompanyID = comp.ID

	// nil スライスは SQL の NULL になり NOT NULL 制約に弾かれる（#30 と同型の正規化）
	if params.RequiredSkills == nil {
		params.RequiredSkills = []string{}
	}

	project, err := u.queries.CreateProject(ctx, params)
	if err != nil {
		return db.Project{}, fmt.Errorf("案件の作成に失敗: %w", err)
	}
	return project, nil
}
