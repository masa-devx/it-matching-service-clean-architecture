# DBマイグレーション（sql-migrate）
# ルートからの実行例: make -C migrations up
# 新規マイグレーション作成: make -C migrations new NAME=create_projects
#
# sql-migrate は api-server/go.mod の tool ディレクティブで版固定し、`go tool` で実行する
# （go install 不要・チーム全員と CI が同じ版になる）。
# `go -C ../api-server` で実行するため cwd は api-server になる。
# dbconfig.yml は絶対パスで渡し、config 内の dir も cwd 基準（../migrations/ddl）にしてある
# ＝ sql-migrate を直接叩かず、必ずこの Makefile を経由すること

SQL_MIGRATE := go -C ../api-server tool sql-migrate
CONFIG := -config=$(CURDIR)/dbconfig.yml

.PHONY: up down reset status new test-up e2e-up

up:
	$(SQL_MIGRATE) up $(CONFIG)

# テストDB（tsunagu_test）へスキーマを適用
test-up:
	$(SQL_MIGRATE) up -env=test $(CONFIG)

# e2e fixture 構築用 DB（tsunagu_e2e）へスキーマを適用（make e2e-dump が使う）
e2e-up:
	$(SQL_MIGRATE) up -env=e2e $(CONFIG)

# down は直前の1件だけ巻き戻す（開発時の手直し用。本番では原則使わない）
down:
	$(SQL_MIGRATE) down $(CONFIG)

# reset は全マイグレーションを巻き戻す（全テーブルDROP＝データ全消し。ローカル専用）
reset:
	$(SQL_MIGRATE) down -limit=0 $(CONFIG)

status:
	$(SQL_MIGRATE) status $(CONFIG)

new:
	$(SQL_MIGRATE) new $(CONFIG) $(NAME)
