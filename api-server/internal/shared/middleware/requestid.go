// Package middleware は HTTP レベルの横断的関心事（request_id・アクセスログ・panic recovery）を置く。
// 認可（shared/auth の StrictMiddleware）とは層が違う: こちらは全ルート共通で mux の外側に巻く
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
)

type requestIDKey struct{}

// validRequestID は受け入れる ID の形式。外部入力をそのままログに書くと
// ログインジェクション（改行や制御文字の混入）の入口になるため、英数とハイフンに限定する
var validRequestID = regexp.MustCompile(`^[A-Za-z0-9-]{8,64}$`)

// RequestID はリクエストごとに ID を context とレスポンスヘッダに載せる。
// 呼び出し元（Next.js やロードバランサー）が X-Request-ID を付けてきた場合は**尊重**し、
// フロントとバックのログを同じ ID で突き合わせられるようにする（無い・不正なら採番）。
// 本番のエラー追跡の生命線: ユーザーからの問い合わせ（ヘッダの値）とサーバーログを突き合わせられる
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if !validRequestID.MatchString(id) {
			id = newRequestID()
		}
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
