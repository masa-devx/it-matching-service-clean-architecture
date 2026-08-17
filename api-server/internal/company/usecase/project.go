// Package usecase は company API の業務ロジックを置く。
// 1つの公開メソッドが1つの業務操作を完結させる。複数テーブルへの書き込みが必要になったら、
// この層がトランザクション境界を持つ（Queries.WithTx で束ねる）。
package usecase

import (
	"context"
	"fmt"

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
// 現状は単発クエリのため薄いが、業務が育つ場所はここ:
// 掲載数の上限チェック・通知の発行・複数テーブル更新（＝トランザクション）などはこの層に足す
func (u *Project) Create(ctx context.Context, params db.CreateProjectParams) (db.Project, error) {
	project, err := u.queries.CreateProject(ctx, params)
	if err != nil {
		return db.Project{}, fmt.Errorf("案件の作成に失敗: %w", err)
	}
	return project, nil
}
