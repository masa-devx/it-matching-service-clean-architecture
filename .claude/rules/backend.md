---
description: Go API / DB を実装・修正するときのルール（Stage 1: フラット構成）
globs:
  - "api/**/*.go"
  - "api/**/*.sql"
---

# バックエンド開発ルール（Go / PostgreSQL）

## 構成（Stage 1）

- `main.go` は「DSN決定 → DB接続 → ルート登録 → 起動」の**組み立てだけ**
- 機能はドメイン語彙のファイルに分ける（`auth.go` / `projects.go` / `applications.go` …）
- 共通ヘルパー（writeJSON / writeError 等）は `response.go` に集約（todo-app の型）
- R1（クリーンアーキ化）までは package main の同一パッケージ分割で運用

## Go の作法

- エラーは `(結果, error)` で返し `if err != nil` を省略しない。ラップは `fmt.Errorf("%w")`
- `rows.Close()` は defer、ループ後の `rows.Err()` を確認
- ハンドラのエラー応答は共通ヘルパー経由。本番想定では `err.Error()` をそのまま返さない

## セキュリティ（todo-app で学んだ型を厳守）

- SQL は**プレースホルダ必須**（文字列連結禁止）
- `SELECT *` 禁止（列を明示）
- **user_id / company_id は検証済みJWTから取得**（リクエスト値を信用しない＝IDOR対策）
- JWT検証では**署名方式（HS256）を必ず確認**。パスワードは bcrypt
- 認可は「ロール × リソース所有」の二重チェック

## DB

- スキーマ変更は **sql-migrate** で管理（`migrations/ddl/*.sql`。作成は `make -C migrations new NAME=xxx`・適用は `make -C migrations up`）。DDLを psql で直接流さない。マイグレーションファイルには必ず Up と Down を書く
- 時刻は `TIMESTAMPTZ`、NULLを許さない列は `NOT NULL`、外部キーには索引
- **状態遷移を持つカラム（status）は、許可される遷移をコード上の1か所（遷移表）で定義**し、テーブル駆動テストを書く（Phase 3 以降）

## テストの書き方

- **テスト関数の上に doc コメントを書く**（実装コードの「whyだけ」ルールとは別扱い）。含める内容:
  - **目的**: このテストが守っている仕様・壊れると何が起きるか
  - **観点**: どのケースを網羅しているか（境界値・異常系・セキュリティ要件など）
  - テストは「仕様の実行可能なドキュメント」なので、読んだ人が何を保証しているか分かる粒度で書く
- テーブル駆動テストを基本形にする（ケース＝データ・検証ロジック＝1か所）
- 境界値は両側から張る（8文字未満/8文字ちょうど など）
- セキュリティ要件（値が漏れない・SQLに埋め込まれない）もテストで固定する
- 教材: 詰まった点・設計判断は `docs/DB.md` / `docs/バックエンド.md` へ
