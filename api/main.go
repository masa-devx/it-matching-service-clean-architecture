package main

import (
	"context"
	"errors"
	"log"
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

	dsn, err := mustEnv("DATABASE_URL")
	if err != nil {
		log.Fatal(err)
	}

	if err := initJWT(); err != nil {
		log.Fatal(err)
	}

	if err := openDB(dsn); err != nil {
		log.Fatal(err)
	}
	// 終了直前のClose失敗は打つ手がないため、エラーは明示的に破棄する
	defer func() { _ = db.Close() }()
	log.Println("DB接続OK")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /signup", handleSignup)
	mux.HandleFunc("POST /login", handleLogin)
	mux.Handle("GET /me", requireAuth(http.HandlerFunc(handleMe)))
	mux.Handle("GET /me/profile", requireAuth(http.HandlerFunc(handleGetProfile)))

	srv := &http.Server{
		Addr:    ":" + cfg.port,
		Handler: loggingMiddleware(corsMiddleware(cfg.webOrigin)(mux)),
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
		log.Printf("起動 → http://localhost:%s  (Ctrl+C で停止)", cfg.port)
		// Shutdown 時の ErrServerClosed は正常終了なのでエラー扱いしない
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("サーバ起動エラー:", err)
		}
	}()

	<-ctx.Done()

	// 新規受付を止め、処理中のリクエストを最大10秒待ってから終了する
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdownエラー: %v", err)
		return
	}
	log.Println("shutdown完了")
}
