# Tsunagu Works 開発コマンド集約
# `make` または `make help` で一覧表示

.DEFAULT_GOAL := help

.PHONY: help dev-api dev-web db-up db-down migrate-up migrate-down migrate-status migrate-new seed test lint build

help: ## コマンド一覧を表示
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

## --- 開発サーバー ---

dev-api: ## Go API を起動（air ホットリロード・:8081）
	cd api && air

dev-web: ## Next.js を起動（:3000）
	cd web && npm run dev

## --- DB ---

db-up: ## PostgreSQL コンテナを起動
	docker compose up -d

db-down: ## PostgreSQL コンテナを停止
	docker compose down

migrate-up: ## 未適用のマイグレーションを適用
	$(MAKE) -C migrations up

migrate-down: ## 直前のマイグレーションを1件巻き戻す（ローカル用）
	$(MAKE) -C migrations down

migrate-status: ## マイグレーションの適用状況を表示
	$(MAKE) -C migrations status

migrate-new: ## 新規マイグレーション作成（例: make migrate-new NAME=create_projects）
	$(MAKE) -C migrations new NAME=$(NAME)

seed: ## 開発用シードデータを投入（2回実行しても安全・資格情報はREADME参照）
	docker compose exec -T db psql -U tsunagu -d tsunagu -v ON_ERROR_STOP=1 < migrations/seed.sql

## --- 品質チェック（CIと同一コマンド） ---

test: ## api / web のテストを実行
	cd api && go test ./...
	cd web && npm run test

lint: ## api / web の lint を実行
	cd api && golangci-lint run ./...
	cd web && npm run lint && npm run format:check

build: ## api / web をビルド（型チェック込み）
	cd api && go build ./...
	cd web && npm run build
