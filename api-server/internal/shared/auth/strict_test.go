package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
)

// テスト用に生成型と同形の関数型を定義（~func 制約により本物の生成型と同様に扱える）
type strictFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error)

// TestNewStrictAuth は認証＋ロール認可のミドルウェアを固定する。
//
// 目的: 保護エンドポイントの門番。壊れると未認証の素通り・ロール越境・公開APIの誤ブロックが起きる。
//
// 観点: 仕様由来の集合に載らない operation の素通り／未認証の401（全パターン同一レスポンス）／
// **ロール違いの403**（401 との区別）／正しいロールの通過（Claims が context に載る）。
func TestNewStrictAuth(t *testing.T) {
	requiredOps := map[string]bool{"Protected": true}

	newMW := func() func(strictFunc, string) strictFunc {
		return auth.NewStrictAuth[strictFunc](testSecret, auth.RoleCompany, requiredOps)
	}

	issue := func(t *testing.T, role string) string {
		token, err := auth.IssueToken(testSecret, 42, role, time.Hour)
		if err != nil {
			t.Fatalf("発行に失敗: %v", err)
		}
		return token
	}

	t.Run("集合に載らない operation は認証なしで素通りする", func(t *testing.T) {
		called := false
		next := func(ctx context.Context, w http.ResponseWriter, r *http.Request, req any) (any, error) {
			called = true
			return nil, nil
		}
		rec := httptest.NewRecorder()
		_, _ = newMW()(next, "Public")(context.Background(), rec, httptest.NewRequest(http.MethodGet, "/x", nil), nil)
		if !called {
			t.Error("公開 operation が誤ってブロックされた")
		}
	})

	t.Run("正しいロールは通過し、Claims が context に載る", func(t *testing.T) {
		var got auth.Claims
		next := func(ctx context.Context, w http.ResponseWriter, r *http.Request, req any) (any, error) {
			claims, ok := auth.ClaimsFrom(ctx)
			if !ok {
				t.Error("Claims が context に無い")
			}
			got = claims
			return nil, nil
		}
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.Header.Set("Authorization", "Bearer "+issue(t, auth.RoleCompany))
		rec := httptest.NewRecorder()
		_, _ = newMW()(next, "Protected")(context.Background(), rec, req, nil)
		if got.UserID != 42 {
			t.Errorf("Claims が一致しない: %+v", got)
		}
	})

	t.Run("ロール違い（talent が company の保護APIへ）は 403", func(t *testing.T) {
		next := func(ctx context.Context, w http.ResponseWriter, r *http.Request, req any) (any, error) {
			t.Error("拒否されるべきリクエストが next に到達した")
			return nil, nil
		}
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		req.Header.Set("Authorization", "Bearer "+issue(t, auth.RoleTalent))
		rec := httptest.NewRecorder()
		_, _ = newMW()(next, "Protected")(context.Background(), rec, req, nil)
		if rec.Code != http.StatusForbidden {
			t.Errorf("403 を期待したが %d", rec.Code)
		}
	})

	// 未認証の全パターンが「同一の401」であること（失敗理由を漏らさない）
	t.Run("未認証は同一の401", func(t *testing.T) {
		rejects := []struct {
			name  string
			setup func(r *http.Request)
		}{
			{name: "ヘッダーなし", setup: func(r *http.Request) {}},
			{name: "Bearer 以外", setup: func(r *http.Request) { r.Header.Set("Authorization", "Basic xxx") }},
			{name: "不正なトークン", setup: func(r *http.Request) { r.Header.Set("Authorization", "Bearer bad.token") }},
			{name: "期限切れ", setup: func(r *http.Request) {
				token, _ := auth.IssueToken(testSecret, 42, auth.RoleCompany, -time.Hour)
				r.Header.Set("Authorization", "Bearer "+token)
			}},
		}

		var bodies []string
		for _, tt := range rejects {
			next := func(ctx context.Context, w http.ResponseWriter, r *http.Request, req any) (any, error) {
				t.Errorf("%s: 拒否されるべきリクエストが next に到達した", tt.name)
				return nil, nil
			}
			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			tt.setup(req)
			rec := httptest.NewRecorder()
			_, _ = newMW()(next, "Protected")(context.Background(), rec, req, nil)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("%s: 401 を期待したが %d", tt.name, rec.Code)
			}
			bodies = append(bodies, rec.Body.String())
		}
		for i := 1; i < len(bodies); i++ {
			if bodies[i] != bodies[0] {
				t.Errorf("拒否レスポンスが同一でない: %q vs %q", bodies[0], bodies[i])
			}
		}
	})
}
