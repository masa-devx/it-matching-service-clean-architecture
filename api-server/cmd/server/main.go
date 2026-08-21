package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/app"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/config"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/infra"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/middleware"
)

// main は起動と停止だけを行う: 設定読込 → DB接続 → 配線（internal/app）→ 起動 → graceful shutdown。
// 配線そのものは app.NewRouter に集約されており、テスト（test/apitest）と同一の組み立てが使われる（#93）
func main() {
	// ロガーは最初に設定する（以降の全ログが JSON 構造化になる）
	slog.SetDefault(middleware.NewLogger())

	if err := run(); err != nil {
		slog.Error("起動に失敗", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// トレース基盤（OTEL_EXPORTER_OTLP_ENDPOINT 未設定なら no-op）。
	// 終了時に flush する: Batcher が溜めている送信前の span を失わないため
	shutdownTracing, err := infra.SetupTracing(ctx, cfg.OTLPEndpoint)
	if err != nil {
		return err
	}
	defer func() {
		flushCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownTracing(flushCtx); err != nil {
			slog.Warn("トレースの flush に失敗", "err", err)
		}
	}()

	pool, err := infra.NewDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	handler, err := app.NewRouter(pool, db.New(pool), cfg.JWTSecret)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second, // Slowloris 対策（ヘッダーを送り切らない接続を切る）
	}

	// graceful shutdown: SIGTERM（Cloud Run のコンテナ入替）/ SIGINT（Ctrl+C）で
	// 新規受付を止め、処理中のリクエストの完了を待ってから終了する。
	// 即死させると入替のたびに誰かのリクエストが切断される
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("api-server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-sigCtx.Done():
	}

	slog.Info("シグナルを受信、graceful shutdown を開始")
	// 完了待ちには上限を設ける（Cloud Run の SIGKILL 猶予より短く）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown に失敗: %w", err)
	}
	slog.Info("shutdown 完了")
	return nil
}
