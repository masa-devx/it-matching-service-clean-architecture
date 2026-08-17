package auth

import (
	"net/http"
	"strings"
)

// Middleware は Bearer トークンを検証し、本人情報（Claims）を context に載せる。
// 失敗はすべて「同一の 401」で返す: ヘッダー無し・改ざん・期限切れを区別して返すと、
// 攻撃者への情報（どこまで合っていたか）になるため（tsunagu-works の型）。
// どのパスに適用するか・ロールの強制は main 側の配線の責務（#31）
func Middleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok {
				unauthorized(w)
				return
			}

			claims, err := VerifyToken(secret, token)
			if err != nil {
				unauthorized(w)
				return
			}

			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"認証が必要です"}`))
}
