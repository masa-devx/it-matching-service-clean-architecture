package main

import (
	"log/slog"
	"os"
	"strings"
)

// initLogger は構造化ログ（log/slog）の既定ロガーを設定する。
// 出力先は常に標準出力にする（ファイルに書かない）: 収集はDocker/クラウド側の仕事という
// 12-Factor の原則に従い、アプリはログの行き先を意識しない
func initLogger(cfg config) {
	opts := &slog.HandlerOptions{Level: parseLevel(cfg.logLevel)}

	var handler slog.Handler
	if strings.EqualFold(cfg.logFormat, "json") {
		// 本番想定: 1行1JSON。ログ基盤がフィールド単位で検索・集計・アラートできる
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// 開発想定: 人が読める key=value 形式
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// 既定ロガーを差し替え、どこからでも slog.Info(...) で書けるようにする
	slog.SetDefault(slog.New(handler))
}

// parseLevel は文字列をログレベルに変換する。不正な値は info 扱い（起動を止めない）
func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
