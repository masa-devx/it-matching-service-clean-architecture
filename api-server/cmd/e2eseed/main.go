// e2eseed は基準世界（test/e2efixture）を専用 DB に構築する。`make e2e-dump` から呼ばれ、
// 直後の pg_dump で dump.sql（生成物）が作られる。
// テスト実行時にこのコマンドは使わない（テストは dump.sql をトランザクション内にロードする）
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/e2efixture"
)

func main() {
	url := os.Getenv("E2E_DATABASE_URL")
	if url == "" {
		// ローカル既定値（開発 DB と同居する専用データベース。パスワードはローカル学習用）
		url = "postgres://tsunagu:tsunagu@localhost:5435/tsunagu_e2e?sslmode=disable"
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		log.Fatalf("e2eseed: DB 接続に失敗: %v", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	if err := e2efixture.Seed(ctx, conn); err != nil {
		log.Fatalf("e2eseed: %v", err)
	}
	fmt.Println("e2eseed: 基準世界を構築しました")
}
