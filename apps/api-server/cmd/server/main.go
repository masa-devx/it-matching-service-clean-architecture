package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Phase 0 の最小サーバー。ルーティングは Phase 0-2 以降 oapi-codegen の生成コードに任せるため、
// ここには手書きのルートを増やさない（health は生成対象外の運用エンドポイントなので例外）
func main() {
	// 環境変数は外部入力として扱い、数値であることを検証してから使う（不正値は fail fast）
	port := 8082
	if v := os.Getenv("PORT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("PORT が数値ではありません: %v", err)
		}
		port = n
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, // Slowloris 対策（ヘッダーを送り切らない接続を切る）
	}

	log.Printf("api-server listening on :%d", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
