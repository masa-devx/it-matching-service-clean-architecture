package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// writeJSON は任意の値をJSONにして指定ステータスで返す
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("レスポンス書き込み失敗:", err)
	}
}

// writeError はエラー詳細をログにのみ出し、クライアントには安全なメッセージだけ返す
// （内部エラーの内容をそのまま返すと、DB構造などの情報漏えいに繋がるため）
func writeError(w http.ResponseWriter, status int, message string, err error) {
	if err != nil {
		log.Printf("[error] %s: %v", message, err)
	}
	writeJSON(w, status, map[string]string{"error": message})
}
