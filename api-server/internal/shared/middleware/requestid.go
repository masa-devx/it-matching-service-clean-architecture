// Package middleware は HTTP レベルの横断的関心事（request_id・アクセスログ・panic recovery）を置く。
// 認可（shared/auth の StrictMiddleware）とは層が違う: こちらは全ルート共通で mux の外側に巻く
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type requestIDKey struct{}

// RequestID はリクエストごとに ID を採番して context とレスポンスヘッダに載せる。
// 本番のエラー追跡の生命線: ユーザーからの問い合わせ（ヘッダの値）とサーバーログを突き合わせられる
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := newRequestID()
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFrom は context から request_id を取り出す（ログハンドラが使う）
func RequestIDFrom(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey{}).(string)
	return id, ok
}

func newRequestID() string {
	// 8バイト（16進16文字）で十分: 衝突耐性より「ログを追える一意性」が目的
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
