package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// .env があれば読み込む。無くてもエラーにしない（本番は実環境変数を使うため）
	_ = godotenv.Load()

	cfg := loadConfig()
	initLogger(cfg)

	dsn, err := mustEnv("DATABASE_URL")
	if err != nil {
		fatal("設定エラー", err)
	}

	if err := initJWT(); err != nil {
		fatal("設定エラー", err)
	}

	if err := openDB(dsn); err != nil {
		fatal("DB接続エラー", err)
	}
	// 終了直前のClose失敗は打つ手がないため、エラーは明示的に破棄する
	defer func() { _ = db.Close() }()
	slog.Info("DB接続OK")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /signup", handleSignup)
	mux.HandleFunc("POST /login", handleLogin)
	mux.Handle("GET /me", requireAuth(http.HandlerFunc(handleMe)))
	mux.Handle("GET /me/profile", requireAuth(http.HandlerFunc(handleGetProfile)))
	mux.Handle("PUT /me/profile", requireAuth(http.HandlerFunc(handlePutProfile)))
	// 案件の掲載は企業のみ（requireAuth の内側で requireRole を重ねる）
	mux.Handle("POST /projects", requireAuth(requireRole(roleCompany)(http.HandlerFunc(handleCreateProject))))
	mux.Handle("GET /projects", requireAuth(http.HandlerFunc(handleListProjects)))
	mux.Handle("GET /projects/{id}", requireAuth(http.HandlerFunc(handleGetProject)))
	// 応募は人材のみ。企業ロールをここで弾くことで自社案件への応募が構造的に起きない
	mux.Handle("POST /projects/{id}/applications",
		requireAuth(requireRole(roleTalent)(http.HandlerFunc(handleCreateApplication))))
	mux.Handle("GET /projects/{id}/applications",
		requireAuth(requireRole(roleCompany)(http.HandlerFunc(handleListProjectApplications))))
	mux.Handle("GET /me/applications",
		requireAuth(requireRole(roleTalent)(http.HandlerFunc(handleListMyApplications))))
	// 選考の状態更新は企業・人材の両方が使う。誰がどの遷移を実行できるかは遷移表で判定するため
	// requireRole は付けない（ロールで入口を塞ぐのではなく、操作ごとに許可する）
	mux.Handle("PATCH /applications/{id}/status",
		requireAuth(http.HandlerFunc(handleUpdateApplicationStatus)))

	srv := &http.Server{
		Addr: ":" + cfg.port,
		// requestID を最外に置き、以降のすべてのログ・レスポンスに相関IDが乗るようにする
		Handler: requestIDMiddleware(recoverMiddleware(loggingMiddleware(corsMiddleware(cfg.webOrigin)(mux)))),
		// Slowloris対策: 接続を握ったまま少しずつ送る攻撃をタイムアウトで切る
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// SIGINT(Ctrl+C)/SIGTERM(コンテナ停止) を受けると ctx が閉じる
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("サーバー起動", "url", "http://localhost:"+cfg.port)
		// Shutdown 時の ErrServerClosed は正常終了なのでエラー扱いしない
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fatal("サーバー起動エラー", err)
		}
	}()

	<-ctx.Done()

	// 新規受付を止め、処理中のリクエストを最大10秒待ってから終了する
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdownエラー", "error", err)
		return
	}
	slog.Info("shutdown完了")
}

// fatal は致命的エラーを構造化ログで出して終了する（log.Fatal の置き換え）
func fatal(msg string, err error) {
	slog.Error(msg, "error", err)
	os.Exit(1)
}
