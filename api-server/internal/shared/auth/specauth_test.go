package auth_test

import (
	"testing"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
	talent "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/talent"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
)

// TestRequiredAuthOps は「認証要否が仕様から導出される」ことを固定する。
//
// 目的: 認証の要否の一次情報を仕様（@useAuth → security 定義）に一本化する。
// 壊れると「仕様では要認証なのに素通り」または「公開のはずが401」になる。
// 仕様に endpoints を足せばこのテストの期待値も仕様から自動で変わる（コードの手書きリスト廃止）。
//
// 観点: 要認証（me / projects作成）と公開（signup / login）の両方を、company / talent の
// 実際の埋め込み仕様（GetSwagger）に対して検証する。
func TestRequiredAuthOps(t *testing.T) {
	t.Run("company: me と ProjectsCreate が必須・signup/login は公開", func(t *testing.T) {
		spec, err := company.GetSpec()
		if err != nil {
			t.Fatalf("埋め込み仕様の取得に失敗: %v", err)
		}

		ops := auth.RequiredAuthOps(spec)

		for _, want := range []string{"AuthMe", "ProjectsCreate"} {
			if !ops[want] {
				t.Errorf("%s が認証必須になっていない", want)
			}
		}
		for _, public := range []string{"AuthSignup", "AuthLogin"} {
			if ops[public] {
				t.Errorf("%s が誤って認証必須になっている", public)
			}
		}
	})

	t.Run("talent: me が必須・signup/login は公開", func(t *testing.T) {
		spec, err := talent.GetSpec()
		if err != nil {
			t.Fatalf("埋め込み仕様の取得に失敗: %v", err)
		}

		ops := auth.RequiredAuthOps(spec)

		if !ops["AuthMe"] {
			t.Error("AuthMe が認証必須になっていない")
		}
		for _, public := range []string{"AuthSignup", "AuthLogin"} {
			if ops[public] {
				t.Errorf("%s が誤って認証必須になっている", public)
			}
		}
	})
}
