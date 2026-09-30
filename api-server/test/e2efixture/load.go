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

// FixtureTables は基準世界のテーブル（FK の依存順）。Makefile の e2e-dump と同じ列挙。
// cmd/e2ereset の TRUNCATE 対象もこれを使う
var FixtureTables = []string{"users", "companies", "talents", "projects", "applications"}

// Load は基準世界をテストのトランザクション内に読み込む。
// ROLLBACK で消えるため後始末は不要で、テストは並列実行できる
func Load(t *testing.T, dbtx db.DBTX) {
	t.Helper()
	if err := LoadInto(context.Background(), dbtx); err != nil {
		t.Fatal(err)
	}
}

// LoadInto は基準世界を dbtx に読み込む本体。API 統合テスト（Load）とブラウザ E2E の
// リセット（cmd/e2ereset）が共用し、「基準世界の正しい読み込み方」を1か所に保つ。
//
// dump は「1行 = 1 INSERT 文」を前提に行単位で実行する（pg_dump --column-inserts の出力形式。
// 基準世界の文字列データに改行を入れないのが運用ルールで、破った場合はここで検出される）
func LoadInto(ctx context.Context, dbtx db.DBTX) error {
	for i, line := range strings.Split(strings.TrimSpace(dumpSQL), "\n") {
		if !strings.HasPrefix(line, "INSERT INTO") || !strings.HasSuffix(line, ");") {
			return fmt.Errorf("dump.sql の %d 行目が INSERT 文の形をしていません（基準世界のデータに改行を入れた可能性）: %.80s", i+1, line)
		}
		if _, err := dbtx.Exec(ctx, line); err != nil {
			return fmt.Errorf("dump.sql の適用に失敗（%d 行目）: %w", i+1, err)
		}
	}

	// dump は ID を明示して INSERT するため、シーケンスが基準世界の ID を追い越していないと
	// 新規 INSERT（API 経由の応募作成など）が ID 重複で失敗する。
	// setval はトランザクションでロールバックされない（非トランザクショナル）ため、
	// 並列テストのシーケンスを巻き戻さないよう GREATEST で「前方にのみ」進める
	for _, table := range FixtureTables {
		q := fmt.Sprintf(`SELECT setval(
			pg_get_serial_sequence('%[1]s', 'id'),
			GREATEST(
				(SELECT COALESCE(max(id), 1) FROM %[1]s),
				COALESCE(pg_sequence_last_value(pg_get_serial_sequence('%[1]s', 'id')), 1)
			))`, table)
		if _, err := dbtx.Exec(ctx, q); err != nil {
			return fmt.Errorf("%s のシーケンス調整に失敗: %w", table, err)
		}
	}
	return nil
}
