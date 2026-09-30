// e2ereset はブラウザ E2E 専用 DB を基準世界の状態に戻す（ADR-0014）。
// Playwright の globalSetup から `make e2e-web-reset` 経由で毎回呼ばれる。
//
// API 統合テストは Tx の ROLLBACK で後始末するが、ブラウザ E2E は実サーバーが COMMIT するため
// 使えない。代わりに「実行前に全消し → 基準世界を読み込み」を1トランザクションで行い、
// 前回の実行の残りや途中失敗に結果が左右されないようにする
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/e2efixture"
)

// requiredDBSuffix は接続先の DB 名に必須の接尾辞。全テーブルを TRUNCATE するコマンドのため、
// 環境変数の設定ミスで開発 DB や本番（Neon）を消さないよう、名前で実行対象を限定する
const requiredDBSuffix = "_e2e_web"

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatalf("e2ereset: %v", err)
	}
	fmt.Println("e2ereset: E2E 用 DB を基準世界にリセットしました")
}

func run(ctx context.Context) error {
	url := os.Getenv("E2E_WEB_DATABASE_URL")
	if url == "" {
		// ローカル既定値。docker-compose.yml にコミット済みの資格情報と同一で、
		// 秘匿情報ではないため gosec の G101 を除外する（helpers/db.go と同じ判断）
		url = "postgres://tsunagu:tsunagu@localhost:5435/tsunagu_e2e_web?sslmode=disable" //nolint:gosec
	}

	cfg, err := pgx.ParseConfig(url)
	if err != nil {
		return fmt.Errorf("接続文字列の解析に失敗: %w", err)
	}
	if !strings.HasSuffix(cfg.Database, requiredDBSuffix) {
		return fmt.Errorf("接続先 DB %q は E2E 専用ではありません（名前が %q で終わる DB だけをリセットできます）", cfg.Database, requiredDBSuffix)
	}

	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("DB 接続に失敗（make db-up を確認してください）: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗: %w", err)
	}
	// Commit 済みなら Rollback は何もしない。途中で失敗したら全消しも取り消され、中途半端な世界を残さない
	defer func() { _ = tx.Rollback(ctx) }()

	// RESTART IDENTITY で ID の採番も初期化し、実行のたびに同じ ID の世界になるようにする。
	// CASCADE は将来このテーブル群を参照するテーブルが増えても TRUNCATE が FK で止まらないため
	tables := make([]string, len(e2efixture.FixtureTables))
	for i, t := range e2efixture.FixtureTables {
		tables[i] = pgx.Identifier{t}.Sanitize()
	}
	if _, err := tx.Exec(ctx, "TRUNCATE "+strings.Join(tables, ", ")+" RESTART IDENTITY CASCADE"); err != nil {
		return fmt.Errorf("TRUNCATE に失敗: %w", err)
	}

	if err := e2efixture.LoadInto(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("コミットに失敗: %w", err)
	}
	return nil
}
