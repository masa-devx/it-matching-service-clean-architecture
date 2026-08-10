// Package infra は外部リソースへの接続（DB・将来はストレージ等）を担う。
package infra

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDB は接続プールを生成し、疎通確認まで行う。
// 起動時に Ping することで「DBに繋がらないまま起動してリクエストで初めて失敗する」を防ぐ（fail fast）
func NewDB(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("DB接続プールの生成に失敗: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("DB疎通確認に失敗: %w", err)
	}
	return pool, nil
}
