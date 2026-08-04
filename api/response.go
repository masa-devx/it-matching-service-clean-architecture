package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// errorResponse はエラー時のレスポンス。request_id を含めることで、
// ユーザーからの「このIDでエラーが出た」という報告だけでログを特定できる
type errorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

// writeJSON は任意の値をJSONにして指定ステータスで返す
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("レスポンス書き込み失敗", "error", err)
	}
}

// writeError はエラー詳細をログにのみ出し、クライアントには安全なメッセージだけ返す
// （内部エラーの内容をそのまま返すと、DB構造などの情報漏えいに繋がるため）。
// ctx から相関IDを取り出し、ログとレスポンスの両方に載せて突き合わせ可能にする
func writeError(ctx context.Context, w http.ResponseWriter, status int, message string, err error) {
	// ステータスでレベルを分ける（4xx=クライアント起因のWarn / 5xx=要対応のError）。
	// err が無い場合もエラー応答の事実は記録する（監視で件数を追えるようにする）
	attrs := []any{"status", status}
	if err != nil {
		attrs = append(attrs, "error", err)
	}
	// 認証後のハンドラでは user_id も残す（誰のリクエストで失敗したかの手がかり）
	if userID, ok := userIDFrom(ctx); ok {
		attrs = append(attrs, "user_id", userID)
	}
	contextLogger(ctx).Log(ctx, levelForStatus(status), message, attrs...)

	writeJSON(w, status, errorResponse{
		Error:     message,
		RequestID: requestIDFrom(ctx),
	})
}
