// Package helpers はテスト共通のセットアップを置く。
package helpers

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
)

// NewTestQueries はテスト用DBへの接続をトランザクションで包んだ Queries を返す。
// テスト終了時に自動で ROLLBACK されるため、テストが何を書き込んでも DB は元通りになる
// ＝テスト間の独立性を「後片付けコード」ではなくトランザクションで保証する（ADR-0008）。
func NewTestQueries(t *testing.T) *db.Queries {
	t.Helper()
	_, queries := NewTestTx(t)
	return queries
}

// NewTestTx はテスト用のトランザクションと、その上の Queries を返す。
// usecase がさらに Begin する場合（SAVEPOINT による入れ子）に、トランザクション自体を
// TxBeginner として渡すために使う（ADR-0008 の入れ子ケース）
func NewTestTx(t *testing.T) (pgx.Tx, *db.Queries) {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		// ローカル学習用の既定値。docker-compose.yml にコミット済みの資格情報と同一で、
		// 秘匿情報ではないため gosec の G101 を除外する
		url = "postgres://tsunagu:tsunagu@localhost:5435/tsunagu_test?sslmode=disable" //nolint:gosec
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatalf("テストDBに接続できません（make db-up && make db-test-setup を実行してください）: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("トランザクションの開始に失敗: %v", err)
	}
	// Cleanup は登録の逆順で実行される: ROLLBACK → Close の順になる
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })

	// pgx.Tx は sqlc の DBTX インターフェースを満たすため、そのまま Queries の土台にできる
	return tx, db.New(tx)
}
