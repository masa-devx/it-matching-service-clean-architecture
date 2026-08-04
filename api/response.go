package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// writeJSON は任意の値をJSONにして指定ステータスで返す
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("レスポンス書き込み失敗", "error", err)
	}
}

// writeError はエラー詳細をログにのみ出し、クライアントには安全なメッセージだけ返す
// （内部エラーの内容をそのまま返すと、DB構造などの情報漏えいに繋がるため）
func writeError(w http.ResponseWriter, status int, message string, err error) {
	// ステータスでレベルを分ける（4xx=クライアント起因のWarn / 5xx=要対応のError）。
	// err が無い場合もエラー応答の事実は記録する（監視で件数を追えるようにする）
	attrs := []any{"status", status}
	if err != nil {
		attrs = append(attrs, "error", err)
	}
	slog.Log(context.Background(), levelForStatus(status), message, attrs...)
	writeJSON(w, status, map[string]string{"error": message})
}
