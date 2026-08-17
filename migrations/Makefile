# DBマイグレーション（sql-migrate）
# ルートからの実行例: make -C migrations up
# 新規マイグレーション作成: make -C migrations new NAME=create_projects

.PHONY: up down reset status new test-up

up:
	sql-migrate up

# テストDB（tsunagu_test）へスキーマを適用
test-up:
	sql-migrate up -env=test

# down は直前の1件だけ巻き戻す（開発時の手直し用。本番では原則使わない）
down:
	sql-migrate down

# reset は全マイグレーションを巻き戻す（全テーブルDROP＝データ全消し。ローカル専用）
reset:
	sql-migrate down -limit=0

status:
	sql-migrate status

new:
	sql-migrate new $(NAME)
