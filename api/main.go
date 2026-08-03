package main

import (
	"log"
	"net/http"

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

	log.Printf("起動 → http://localhost:%s  (Ctrl+C で停止)", cfg.port)
	//nolint:gosec // タイムアウト設定可能な http.Server への移行は #29 で対応
	if err := http.ListenAndServe(":"+cfg.port, corsMiddleware(cfg.webOrigin)(mux)); err != nil {
		log.Fatal("サーバ起動エラー:", err)
	}
}
