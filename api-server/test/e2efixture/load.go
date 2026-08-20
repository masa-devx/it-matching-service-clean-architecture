package e2efixture

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
)

// dump.sql は make e2e-dump が生成する（手編集禁止・1行=1 INSERT 文）。
// バイナリに埋め込むことで、テストの実行場所（CI・ローカル）に依らずロードできる
//
//go:embed dump.sql
var dumpSQL string

// fixtureTables は基準世界のテーブル（FK の依存順）。Makefile の e2e-dump と同じ列挙
var fixtureTables = []string{"users", "companies", "talents", "projects", "applications"}

// Load は基準世界をテストのトランザクション内に読み込む。
// ROLLBACK で消えるため後始末は不要で、テストは並列実行できる。
//
// dump は「1行 = 1 INSERT 文」を前提に行単位で実行する（pg_dump --column-inserts の出力形式。
// 基準世界の文字列データに改行を入れないのが運用ルールで、破った場合はここで検出される）
func Load(t *testing.T, dbtx db.DBTX) {
	t.Helper()
	ctx := context.Background()

	for i, line := range strings.Split(strings.TrimSpace(dumpSQL), "\n") {
		if !strings.HasPrefix(line, "INSERT INTO") || !strings.HasSuffix(line, ");") {
			t.Fatalf("dump.sql の %d 行目が INSERT 文の形をしていません（基準世界のデータに改行を入れた可能性）: %.80s", i+1, line)
		}
		if _, err := dbtx.Exec(ctx, line); err != nil {
			t.Fatalf("dump.sql の適用に失敗（%d 行目）: %v", i+1, err)
		}
	}

	// dump は ID を明示して INSERT するため、シーケンスが基準世界の ID を追い越していないと
	// テスト内の新規 INSERT（API 経由の応募作成など）が ID 重複で失敗する。
	// setval はトランザクションでロールバックされない（非トランザクショナル）ため、
	// 並列テストのシーケンスを巻き戻さないよう GREATEST で「前方にのみ」進める
	for _, table := range fixtureTables {
		q := fmt.Sprintf(`SELECT setval(
			pg_get_serial_sequence('%[1]s', 'id'),
			GREATEST(
				(SELECT COALESCE(max(id), 1) FROM %[1]s),
				COALESCE(pg_sequence_last_value(pg_get_serial_sequence('%[1]s', 'id')), 1)
			))`, table)
		if _, err := dbtx.Exec(ctx, q); err != nil {
			t.Fatalf("%s のシーケンス調整に失敗: %v", table, err)
		}
	}
}
