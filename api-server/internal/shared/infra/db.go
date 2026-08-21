// Package infra は外部リソースへの接続（DB・将来はストレージ等）を担う。
package infra

import (
	"context"
	"fmt"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewDB は接続プールを生成し、疎通確認まで行う。
// 起動時に Ping することで「DBに繋がらないまま起動してリクエストで初めて失敗する」を防ぐ（fail fast）
func NewDB(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("DB接続設定の解析に失敗: %w", err)
	}
	// SQL の実行を span として記録する（どのクエリに時間がかかったかが trace の木構造で見える）。
	// TracerProvider が no-op なら記録されない＝トレース無効時のコストはほぼゼロ
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("DB接続プールの生成に失敗: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("DB疎通確認に失敗: %w", err)
	}
	return pool, nil
}
