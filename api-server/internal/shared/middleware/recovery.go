package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recovery はハンドラの panic を 500 に変換する（1リクエストの事故でプロセスを道連れにしない）。
// スタックトレースはログのみ・クライアントには一般文言（内部情報を漏らさない）
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				slog.ErrorContext(r.Context(), "panic recovered",
					"panic", p,
					"stack", string(debug.Stack()),
				)
				// レスポンス書き込み済みの panic では二重 WriteHeader の警告が出るが、
				// その場合はもう手遅れなので許容する
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"サーバー内部でエラーが発生しました"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
