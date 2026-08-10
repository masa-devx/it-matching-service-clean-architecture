# 開発コマンド集約（アプリの起動・ビルドは pnpm / turbo 側: `pnpm dev` / `pnpm turbo build`）
# `make` または `make help` で一覧表示

.DEFAULT_GOAL := help

.PHONY: help db-up db-down migrate-up migrate-down migrate-status migrate-new seed seed-large perf

help: ## コマンド一覧を表示
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

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

seed: ## 開発用シードデータを投入（2回実行しても安全）
	docker compose exec -T db psql -U tsunagu -d tsunagu -v ON_ERROR_STOP=1 < migrations/seed.sql

seed-large: ## 性能計測用に案件5万件を投入（ローカル専用）
	time docker compose exec -T db psql -U tsunagu -d tsunagu -v ON_ERROR_STOP=1 < migrations/seed_large.sql

## --- 性能計測 ---

perf: ## DBクエリの実行計画を計測（EXPLAIN ANALYZE・要 make seed-large）
	docker compose exec -T db psql -U tsunagu -d tsunagu -v ON_ERROR_STOP=1 < migrations/perf_queries.sql
