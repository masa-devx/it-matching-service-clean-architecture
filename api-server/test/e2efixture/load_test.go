package e2efixture_test

import (
	"context"
	"fmt"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/e2efixture"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/factories"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/helpers"
)

// TestLoad は基準世界の読み込みを固定する。
//
// 目的: dump.sql（生成物）が現在のスキーマと整合し、API 統合テスト（#93）の前提となる
// 「全状態の揃った世界」がトランザクション内に再現できることを保証する。
// 壊れると #93 の全テストが前提から崩れる。
//
// 観点: 全テーブルの件数／応募6状態が ids.go の定数どおり存在する／
// Password 定数と seed のハッシュが対応している（signin できる前提）／
// 読み込み後の新規 INSERT が基準世界の ID と衝突しない（シーケンス調整）。
func TestLoad(t *testing.T) {
	ctx := context.Background()
	tx, queries := helpers.NewTestTx(t)
	e2efixture.Load(t, tx)

	counts := map[string]int{"users": 4, "companies": 2, "talents": 2, "projects": 5, "applications": 6}
	for table, want := range counts {
		var got int
		if err := tx.QueryRow(ctx, fmt.Sprintf("SELECT count(*) FROM %s", table)).Scan(&got); err != nil {
			t.Fatalf("%s の件数取得に失敗: %v", table, err)
		}
		if got != want {
			t.Errorf("%s: %d 件を期待したが %d 件", table, want, got)
		}
	}

	statuses := map[int64]string{
		e2efixture.Applications.Applied:   "applied",
		e2efixture.Applications.Offered:   "offered",
		e2efixture.Applications.Accepted:  "accepted",
		e2efixture.Applications.Rejected:  "rejected",
		e2efixture.Applications.Withdrawn: "withdrawn",
		e2efixture.Applications.Declined:  "declined",
	}
	for id, want := range statuses {
		var got string
		if err := tx.QueryRow(ctx, "SELECT status FROM applications WHERE id = $1", id).Scan(&got); err != nil {
			t.Fatalf("応募 %d の取得に失敗: %v", id, err)
		}
		if got != want {
			t.Errorf("応募 %d: status %s を期待したが %s", id, want, got)
		}
	}

	// Password 定数で本物の signin ができること（ハッシュ定数とのペアが正しいこと）を固定する
	var hash string
	if err := tx.QueryRow(ctx, "SELECT password_hash FROM users WHERE id = $1", e2efixture.CompanyA.UserID).Scan(&hash); err != nil {
		t.Fatalf("password_hash の取得に失敗: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(e2efixture.Password)); err != nil {
		t.Errorf("Password 定数がハッシュと一致しない（seed.go の passwordHash を更新し忘れ？）: %v", err)
	}

	// 新規 INSERT が基準世界と衝突しないこと（シーケンス調整の検証）
	user, err := queries.CreateUser(ctx, factories.CreateUserParams())
	if err != nil {
		t.Fatalf("基準世界読み込み後の新規ユーザー作成に失敗（ID 衝突の疑い）: %v", err)
	}
	if user.ID <= e2efixture.TalentB.UserID {
		t.Errorf("新規ユーザーの ID %d が基準世界の範囲（〜%d）と重なっている", user.ID, e2efixture.TalentB.UserID)
	}
}

// TestLoadParallel は基準世界が並列テストで共存できることを固定する。
//
// 目的: fixture 方式の利点「後始末不要・並列実行可」が実際に成り立つことを保証する
// （参考にした実務の dump 方式は全テーブル削除の後始末が必要で直列実行を強制されていた。
// 同じ轍を踏んでいないことをテストで示す）。
//
// 観点: 2つの並列テストが同時に Load し、それぞれが同じ固定 ID の世界を見え、
// 新規 INSERT しても互いに干渉しないこと。
func TestLoadParallel(t *testing.T) {
	for i := range 2 {
		t.Run(fmt.Sprintf("world-%d", i), func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			tx, queries := helpers.NewTestTx(t)
			e2efixture.Load(t, tx)

			var status string
			if err := tx.QueryRow(ctx, "SELECT status FROM applications WHERE id = $1", e2efixture.Applications.Accepted).Scan(&status); err != nil {
				t.Fatalf("応募の取得に失敗: %v", err)
			}
			if status != "accepted" {
				t.Errorf("accepted を期待したが %s", status)
			}

			if _, err := queries.CreateUser(ctx, factories.CreateUserParams()); err != nil {
				t.Errorf("並列下の新規ユーザー作成に失敗: %v", err)
			}
		})
	}
}
