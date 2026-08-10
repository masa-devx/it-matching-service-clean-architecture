---
description: Go API / DB を実装・修正するときのルール（出発点=フラット構成／新構成=設計プラン）
globs:
  - "api/**/*.go"
  - "api/**/*.sql"
  - "apps/api-server/**/*.go"
  - "apps/api-server/**/*.sql"
---

# バックエンド開発ルール（Go / PostgreSQL）

## 構成

**出発点コード（`api/`）を触る場合**

- `main.go` は「DSN決定 → DB接続 → ルート登録 → 起動」の**組み立てだけ**
- 機能はドメイン語彙のファイルに分ける（`auth.go` / `projects.go` / `applications.go` …）
- 共通ヘルパー（writeJSON / writeError 等）は `response.go` に集約（todo-app の型）
- package main の同一パッケージ分割のまま維持する（層分けは新構成側で行う。出発点は Before として保つ）

**新構成（`apps/api-server/`）を作る・触る場合**

- `docs/後継リポジトリ設計プラン.md` §4〜7 を正とする: `internal/{company,talent,shared}` の**視点→層**分割
- 責務: handler=生成IFの実装・詰め替えのみ／usecase=業務ロジック・**トランザクション境界**／`shared/domain`=遷移表・不変条件（DBにもHTTPにも依存しない）
- SQL は `queries/*.sql` に書いて sqlc で生成。**repository 層・`Queries` の薄皮ラップを作らない**
- **生成物（`generated/`）は編集しない**。仕様（`.tsp` / `queries/`）を直して再生成する
- `internal/company` ⇔ `internal/talent` の相互 import 禁止・`shared/domain` から `infra`/`generated` の import 禁止（depguard で強制）

## Go の作法

- エラーは `(結果, error)` で返し `if err != nil` を省略しない。ラップは `fmt.Errorf("%w")`
- `rows.Close()` は defer、ループ後の `rows.Err()` を確認
- ハンドラのエラー応答は共通ヘルパー経由。本番想定では `err.Error()` をそのまま返さない

## セキュリティ（todo-app で学んだ型を厳守）

- SQL は**プレースホルダ必須**（文字列連結禁止）
- `SELECT *` 禁止（列を明示）
- **user_id / company_id は検証済みJWTから取得**（リクエスト値を信用しない＝IDOR対策）
- JWT検証では**署名方式（HS256）を必ず確認**。パスワードは bcrypt
- 認可は「ロール × リソース所有」の二重チェック。**新構成ではロール判定をハンドラに書かない**（パスプレフィックス × ミドルウェアで一律）。所有チェックは SQL の WHERE に埋め込む

## DB

- スキーマ変更は **sql-migrate** で管理（`migrations/ddl/*.sql`。作成は `make -C migrations new NAME=xxx`・適用は `make -C migrations up`）。DDLを psql で直接流さない。マイグレーションファイルには必ず Up と Down を書く
- 時刻は `TIMESTAMPTZ`、NULLを許さない列は `NOT NULL`、外部キーには索引
- **状態遷移を持つカラム（status）は、許可される遷移をコード上の1か所（遷移表）で定義**し、テーブル駆動テストを書く（新構成では遷移表は `shared/domain` に置く）

## テストの書き方

- **テスト関数の上に doc コメントを書く**（実装コードの「whyだけ」ルールとは別扱い）。含める内容:
  - **目的**: このテストが守っている仕様・壊れると何が起きるか
  - **観点**: どのケースを網羅しているか（境界値・異常系・セキュリティ要件など）
  - テストは「仕様の実行可能なドキュメント」なので、読んだ人が何を保証しているか分かる粒度で書く
- テーブル駆動テストを基本形にする（ケース＝データ・検証ロジック＝1か所）
- **新構成の usecase テストは go-txdb ＋ factories（実DB・モックしない）**。テスト用DBは同一インスタンスの別データベース（`_test` サフィックス）
- 境界値は両側から張る（8文字未満/8文字ちょうど など）
- セキュリティ要件（値が漏れない・SQLに埋め込まれない）もテストで固定する
- 教材: 詰まった点・設計判断は `docs/DB.md` / `docs/バックエンド.md` へ
