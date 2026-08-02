package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Postgresドライバ（database/sql から "pgx" 名で利用）
)

// db はアプリ全体で共有する接続プール
var db *sql.DB

// openDB は接続プールを生成し、疎通確認まで行う
func openDB(dsn string) error {
	var err error
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("DB設定エラー: %w", err)
	}

	// 接続プールの上限。無制限だと高負荷時に接続が枯渇してDBを巻き込むため明示する
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("DB接続失敗（docker compose up -d 済みか確認）: %w", err)
	}
	return nil
}
