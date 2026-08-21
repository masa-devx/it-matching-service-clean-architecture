// Package config は環境変数の読み取りを一元化する。ハードコード禁止・読み取りはここだけ。
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        int
	DatabaseURL string
	JWTSecret   []byte
	// OTLPEndpoint はトレースの送信先（例 http://localhost:4318）。空なら送信しない（no-op）
	OTLPEndpoint string
}

// Load は環境変数から設定を組み立てる。
// 接続先など「間違った値で動くと危険な設定」は必須（fail fast）、
// ポートなど「デフォルトで安全に動く設定」は任意＋既定値、と使い分ける
func Load() (Config, error) {
	// .env があれば読み込む。無くてもエラーにしない（CI・本番は実環境変数で渡すため）
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL が設定されていません（cp .env.example .env を確認してください）")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET が設定されていません（cp .env.example .env を確認してください）")
	}

	// 環境変数は外部入力として扱い、数値であることを検証してから使う
	port := 8082
	if v := os.Getenv("PORT"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("PORT が数値ではありません: %w", err)
		}
		port = n
	}

	return Config{
		Port:         port,
		DatabaseURL:  dbURL,
		JWTSecret:    []byte(jwtSecret),
		OTLPEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	}, nil
}
