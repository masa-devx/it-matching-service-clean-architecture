package main

import (
	"context"
	"net/http"
	"strings"
)

// ctxKey は context に値を載せるときの専用キー型。
// string 等の組み込み型をキーにすると他ライブラリのキーと衝突しうるため、
// 非公開の独自型にして衝突を構造的に防ぐ（context パッケージの推奨イディオム）
type ctxKey int

const ctxKeyUserID ctxKey = iota

// requireAuth は Authorization: Bearer の JWT を検証し、
// userID を context に載せて次のハンドラへ渡す。失敗理由は問わず一律 401
func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || tokenString == "" {
			writeError(w, http.StatusUnauthorized, "認証が必要です", nil)
			return
		}

		userID, err := parseToken(tokenString)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "認証が必要です", err)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyUserID, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// userIDFrom は requireAuth が context に格納した userID を取り出す。
// user_id は必ずここから取得する（リクエスト値を信用しない＝IDOR対策）
func userIDFrom(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(ctxKeyUserID).(int64)
	return id, ok
}

// corsMiddleware はブラウザからのクロスオリジンアクセスを origin（WEB_ORIGIN）にだけ許可する
// ミドルウェアを返す。許可リスト方式（"*" にしない）: Cookie 認証では "*" が仕様上使えない上、
// どのサイトからもAPIを叩ける状態は避ける
func corsMiddleware(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Origin ごとに応答が変わることをキャッシュに伝える
			w.Header().Set("Vary", "Origin")

			if r.Header.Get("Origin") == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// プリフライト（OPTIONS）はここで完結させ、ルーティングに渡さない
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
