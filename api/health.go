package main

import (
	"context"
	"net/http"
	"time"
)

// GET /health … 死活監視用エンドポイント。DBの疎通まで含めて確認する
func handleHealth(w http.ResponseWriter, r *http.Request) {
	// DBが応答しない場合に無限に待たないよう、タイムアウトを切る
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		writeError(r.Context(), w, http.StatusServiceUnavailable, "database unavailable", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
