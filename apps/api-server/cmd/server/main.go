package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

// Phase 0 の最小サーバー。ルーティングは Phase 0-2 以降 oapi-codegen の生成コードに任せるため、
// ここには手書きのルートを増やさない（health は生成対象外の運用エンドポイントなので例外）
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second, // Slowloris 対策（ヘッダーを送り切らない接続を切る）
	}

	log.Printf("api-server listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
