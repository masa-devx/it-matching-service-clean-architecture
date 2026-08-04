package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// ctxKey は context に値を載せるときの専用キー型。
// string 等の組み込み型をキーにすると他ライブラリのキーと衝突しうるため、
// 非公開の独自型にして衝突を構造的に防ぐ（context パッケージの推奨イディオム）
type ctxKey int

const (
	ctxKeyUserID ctxKey = iota
	ctxKeyRequestID
)

// requestIDHeader は相関IDの受け渡しに使う事実上の標準ヘッダ名
const requestIDHeader = "X-Request-ID"

// requestIDMiddleware はリクエストごとに一意のIDを発行し、context とレスポンスヘッダに載せる。
// これによりアクセスログ・エラーログ・クライアントの3者を同じIDで突き合わせできる。
// 受信ヘッダに X-Request-ID があれば尊重する（ロードバランサやフロントが採番した値を引き継ぐ）
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := sanitizeRequestID(r.Header.Get(requestIDHeader))
		if id == "" {
			id = newRequestID()
		}

		w.Header().Set(requestIDHeader, id)
		ctx := context.WithValue(r.Context(), ctxKeyRequestID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requestIDFrom は context から相関IDを取り出す（未設定なら空文字）
func requestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(ctxKeyRequestID).(string)
	return id
}

// newRequestID は衝突しない一意なIDを生成する。
// UUIDの版番号や書式は相関IDに不要なため、crypto/rand の16バイトを hex 化するだけにする
func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand の失敗は現実にはほぼ起きない。IDが無くても処理は続ける
		return ""
	}
	return hex.EncodeToString(b[:])
}

// sanitizeRequestID は外部から渡されたIDを検証する。
// ログやレスポンスヘッダに載せる値なので、長さと文字種を制限して汚染を防ぐ
func sanitizeRequestID(id string) string {
	if len(id) == 0 || len(id) > 64 {
		return ""
	}
	for _, c := range id {
		isAllowed := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_'
		if !isAllowed {
			return ""
		}
	}
	return id
}

// requireAuth は Authorization: Bearer の JWT を検証し、
// userID を context に載せて次のハンドラへ渡す。失敗理由は問わず一律 401
func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || tokenString == "" {
			writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", nil)
			return
		}

		userID, err := parseToken(tokenString)
		if err != nil {
			writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", err)
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

// statusRecorder は ResponseWriter を包み、ハンドラが書いたステータスコードを記録する。
// ResponseWriter 自体には「何を書いたか」を後から知る手段がないため、書く瞬間に横取りする
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// loggingMiddleware は全リクエストを「メソッド パス ステータス 所要時間」の1行で記録する。
// 最外層に置くこと（CORSプリフライトや認証で弾かれたリクエストも記録対象にする）
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// WriteHeader が呼ばれないまま Write されたときの既定は 200
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		// 属性として渡すため、フォーマット文字列への埋め込みが無い＝
		// ログインジェクション（改行での偽ログ行の捏造）が構造的に起きない
		contextLogger(r.Context()).LogAttrs(r.Context(), levelForStatus(rec.status), "request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
	})
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
