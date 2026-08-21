# 開発コマンド集約（アプリの起動・ビルドは pnpm / turbo 側: `pnpm dev` / `pnpm turbo build`）
# `make` または `make help` で一覧表示

.DEFAULT_GOAL := help

.PHONY: help db-up db-down migrate-up migrate-down migrate-status migrate-new db-test-setup seed e2e-dump test-api trace-up trace-down dev dev-api dev-web

help: ## コマンド一覧を表示
	@grep -E '^[a-zA-Z0-9_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

## --- DB ---

db-up: ## PostgreSQL コンテナを起動（:5435）
	docker compose up -d db

db-down: ## コンテナを停止
	docker compose down

migrate-up: ## 未適用のマイグレーションを適用
	$(MAKE) -C migrations up

migrate-down: ## 直前のマイグレーションを1件巻き戻す（ローカル用）
	$(MAKE) -C migrations down

migrate-status: ## マイグレーションの適用状況を表示
	$(MAKE) -C migrations status

migrate-new: ## 新規マイグレーション作成（例: make migrate-new NAME=create_projects）
	$(MAKE) -C migrations new NAME=$(NAME)

seed: ## 開発用シードデータを投入（冪等・要 make db-up + migrate-up）
	docker compose exec -T db psql -U tsunagu -d tsunagu -v ON_ERROR_STOP=1 < migrations/seed.sql

## --- テストDB（実DBテスト用） ---

db-test-setup: ## テストDB（tsunagu_test）を作成しスキーマを適用（冪等・要 make db-up）
	docker compose exec -T db psql -U tsunagu -d tsunagu -tc "SELECT 1 FROM pg_database WHERE datname = 'tsunagu_test'" | grep -q 1 || \
		docker compose exec -T db psql -U tsunagu -d tsunagu -c "CREATE DATABASE tsunagu_test"
	$(MAKE) -C migrations test-up


## --- e2e fixture（API 統合テストの基準世界） ---

# DB コンテナ内でコマンドを実行する方法。ローカルは compose、CI はサービスコンテナの docker exec に
# 差し替える（backend-ci が DB_EXEC を上書き）＝ dump 生成手順の一次情報をこの1か所に保つ
DB_EXEC ?= docker compose exec -T db

# dump.sql は生成物（手編集禁止）。世界を変えるときは seed.go を変えてこのターゲットで再生成する。
# テーブルを明示列挙するのは、pg_dump のテーブル順（カタログ順）が FK の依存順と一致する保証がないため
e2e-dump: ## e2e fixture の dump.sql を再生成（一次情報は api-server/test/e2efixture/seed.go・要 make db-up）
	$(DB_EXEC) dropdb -U tsunagu --if-exists tsunagu_e2e
	$(DB_EXEC) createdb -U tsunagu tsunagu_e2e
	$(MAKE) -C migrations e2e-up
	cd api-server && go run ./cmd/e2eseed
	@for t in users companies talents projects applications; do \
		$(DB_EXEC) pg_dump -U tsunagu --data-only --column-inserts -t public.$$t tsunagu_e2e | grep -E '^INSERT INTO'; \
	done > api-server/test/e2efixture/dump.sql
	@test -s api-server/test/e2efixture/dump.sql || { echo "dump.sql が空です（pg_dump 失敗の可能性）"; exit 1; }
	@echo "生成完了: api-server/test/e2efixture/dump.sql（$$(wc -l < api-server/test/e2efixture/dump.sql | tr -d ' ') 行）"

## --- テスト実行（人間の目のためのローカル用。CI は素の go test を turbo 経由で使う） ---

test-api: ## api の全テストを Jest 風の見やすい表示で実行（gotestsum・要 make db-up + db-test-setup）
	cd api-server && go tool gotestsum --format testname --format-hide-empty-pkg -- -count=1 ./...

## --- トレース（ローカル可視化・任意） ---

trace-up: ## Jaeger を起動（UI: http://localhost:16686・OTLP: :4318）
	docker compose --profile trace up -d jaeger

trace-down: ## Jaeger を停止
	docker compose --profile trace down jaeger 2>/dev/null || docker compose stop jaeger

## --- アプリ起動（フォアグラウンド実行・Ctrl+C で停止・要 make db-up） ---

dev: ## バックエンド + フロントエンドを並列起動（= pnpm dev）
	pnpm dev

dev-api: ## バックエンドのみ起動（Go API :8082）
	pnpm turbo dev --filter=api-server

dev-web: ## フロントエンドのみ起動（Next.js :3001・API は別途 dev-api か本番向き先が必要）
	pnpm turbo dev --filter=web
