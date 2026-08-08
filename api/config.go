package main

import (
	"fmt"
	"os"
)

// config はアプリの起動設定。環境変数の読み取りはここに集約する
type config struct {
	port      string // APIの待ち受けポート
	webOrigin string // CORSで許可するフロントの出所
	logFormat string // ログ形式（text=開発 / json=本番想定）
	logLevel  string // ログレベル（debug / info / warn / error）
}

// loadConfig は環境変数から設定を組み立てる。
// 接続先など「間違った値で動くと危険な設定」は必須（呼び出し側でfatal）、
// ポートなど「デフォルトで安全に動く設定」は任意＋既定値、と使い分ける
func loadConfig() config {
	return config{
		port:      envOr("PORT", "8082"),
		webOrigin: envOr("WEB_ORIGIN", "http://localhost:3001"),
		logFormat: envOr("LOG_FORMAT", "text"),
		logLevel:  envOr("LOG_LEVEL", "info"),
	}
}

// envOr は環境変数が未設定・空のとき fallback を返す
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// mustEnv は必須の環境変数を読む。未設定は設定ミスとして起動を止める（fail fast）
func mustEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("%s が設定されていません（cp .env.example .env を確認してください）", key)
	}
	return v, nil
}
