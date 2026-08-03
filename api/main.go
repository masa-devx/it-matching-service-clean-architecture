package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// .env があれば読み込む。無くてもエラーにしない（本番は実環境変数を使うため）
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL が設定されていません（cp .env.example .env を実行してください）")
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

	log.Println("起動 → http://localhost:8081  (Ctrl+C で停止)")
	//nolint:gosec // タイムアウト設定可能な http.Server への移行は #29 で対応
	if err := http.ListenAndServe(":8081", corsMiddleware(mux)); err != nil {
		log.Fatal("サーバ起動エラー:", err)
	}
}
