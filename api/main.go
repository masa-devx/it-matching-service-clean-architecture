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
	defer db.Close()
	log.Println("DB接続OK")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /signup", handleSignup)
	mux.HandleFunc("POST /login", handleLogin)
	mux.Handle("GET /me", requireAuth(http.HandlerFunc(handleMe)))

	log.Println("起動 → http://localhost:8081  (Ctrl+C で停止)")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatal("サーバ起動エラー:", err)
	}
}
