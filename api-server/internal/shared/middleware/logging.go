package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// NewLogger は JSON 構造化ログの logger を作る（Cloud Run のログビューアが1行1JSONを構造化して表示する）。
// contextHandler を挟むことで、slog.ErrorContext(ctx, ...) を呼ぶだけで
// request_id が自動付与される＝呼び出し側は ID を引数で回さない
func NewLogger() *slog.Logger {
	return slog.New(contextHandler{Handler: slog.NewJSONHandler(os.Stdout, nil)})
}

// contextHandler は context から request_id を拾ってレコードに足すデコレータ
type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, record slog.Record) error {
	if id, ok := RequestIDFrom(ctx); ok {
		record.AddAttrs(slog.String("request_id", id))
	}
	// トレースが有効なリクエストならログにも trace_id を載せる
	// （Jaeger 上の trace と grep したログを同じ ID で突き合わせるための「合流点」）
	if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
		record.AddAttrs(slog.String("trace_id", span.TraceID().String()))
	}
	return h.Handler.Handle(ctx, record)
}

// WithAttrs / WithGroup でも自分（デコレータ）を保つ（内側だけ差し替えるとラップが外れる）
func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{Handler: h.Handler.WithGroup(name)}
}

// AccessLog は全リクエストの完了ログ（method / path / status / 所要時間）を1行出す
func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.InfoContext(r.Context(), "request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// statusRecorder はレスポンスのステータスコードを横取りする
// （http.ResponseWriter は書き込んだ status を後から読めないため）
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
