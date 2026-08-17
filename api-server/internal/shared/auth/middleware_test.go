package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
)

// TestMiddleware は認証ミドルウェアの通過と拒否を実ミドルウェア込みで固定する。
//
// 目的: 保護エンドポイントの門番。壊れると未認証アクセスが素通りする/正当なユーザーが弾かれる。
//
// 観点: 有効トークンの通過（context に Claims が載る）と、拒否パターンの網羅。
// セキュリティ要件として「全ての拒否が同一の 401 レスポンス」であること
// （文言の違いで失敗理由を推測させない）もテストで固定する。
func TestMiddleware(t *testing.T) {
	newHandler := func(t *testing.T, gotClaims *auth.Claims) http.Handler {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.ClaimsFrom(r.Context())
			if !ok {
				t.Error("context に Claims が載っていない")
			}
			*gotClaims = claims
			w.WriteHeader(http.StatusOK)
		})
		return auth.Middleware(testSecret)(next)
	}

	t.Run("有効なトークンは通過し、Claims が context に載る", func(t *testing.T) {
		token, err := auth.IssueToken(testSecret, 42, auth.RoleTalent, time.Hour)
		if err != nil {
			t.Fatalf("発行に失敗: %v", err)
		}

		var got auth.Claims
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		newHandler(t, &got).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("200 を期待したが %d", rec.Code)
		}
		if got.UserID != 42 || got.Role != auth.RoleTalent {
			t.Errorf("Claims が一致しない: %+v", got)
		}
	})

	// 拒否パターン: すべて同一の 401 になること
	rejects := []struct {
		name  string
		setup func(r *http.Request)
	}{
		{name: "Authorization ヘッダーなし", setup: func(r *http.Request) {}},
		{name: "Bearer 以外のスキーム", setup: func(r *http.Request) {
			r.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		}},
		{name: "不正なトークン", setup: func(r *http.Request) {
			r.Header.Set("Authorization", "Bearer invalid.token.value")
		}},
		{name: "期限切れトークン", setup: func(r *http.Request) {
			token, _ := auth.IssueToken(testSecret, 42, auth.RoleTalent, -time.Hour)
			r.Header.Set("Authorization", "Bearer "+token)
		}},
	}

	var bodies []string
	for _, tt := range rejects {
		t.Run("拒否: "+tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("拒否されるべきリクエストが next に到達した")
			})
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			tt.setup(req)
			rec := httptest.NewRecorder()

			auth.Middleware(testSecret)(next).ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("401 を期待したが %d", rec.Code)
			}
			bodies = append(bodies, rec.Body.String())
		})
	}

	t.Run("全ての拒否が同一のレスポンス（失敗理由を漏らさない）", func(t *testing.T) {
		for i := 1; i < len(bodies); i++ {
			if bodies[i] != bodies[0] {
				t.Errorf("拒否レスポンスが同一でない: %q vs %q", bodies[0], bodies[i])
			}
		}
	})
}
