// Package main はマイグレーション実行コマンド（Cloud Run Job の実体）。
//
// API 起動時にマイグレーションを流さない理由: Cloud Run は複数インスタンスが
// 同時に起動しうるため、起動時実行だと同じ DDL が並走して競合する。
// 「デプロイパイプラインの中で Job として1回だけ実行する」が正しい置き場所
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	// database/sql に pgx ドライバを登録する（sql-migrate は database/sql を要求。
	// アプリ本体は pgx ネイティブだが、ここだけ stdlib アダプタで橋渡しする）
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/middleware"
)

func main() {
	// Job のログも構造化する（Cloud Run Job のログビューアで server と同じ形式）
	slog.SetDefault(middleware.NewLogger())

	if err := run(); err != nil {
		slog.Error("マイグレーションに失敗", "err", err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return errors.New("DATABASE_URL が設定されていません")
	}

	// 既定はイメージ内のパス（Dockerfile の COPY 先）。ローカル実行時は上書きする
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "/migrations/ddl"
	}

	db, err := sql.Open("pgx", url)
	if err != nil {
		return fmt.Errorf("DB接続の初期化に失敗: %w", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("DB疎通に失敗: %w", err)
	}

	// sql-migrate は適用済みを gorp_migrations で管理するため、何度実行しても冪等
	// （未適用ゼロなら applied=0 で正常終了 = Job のリトライにも安全）
	applied, err := migrate.Exec(db, "postgres", &migrate.FileMigrationSource{Dir: dir}, migrate.Up)
	if err != nil {
		return fmt.Errorf("マイグレーション適用に失敗: %w", err)
	}

	slog.Info("マイグレーション完了", "applied", applied, "dir", dir)
	return nil
}
