# Tsunagu Works

Go（net/http）+ Next.js（App Router）+ PostgreSQL のフルスタックWebアプリケーション。
JWT + httpOnly Cookie の認証・BFF・マイグレーション・CI・テスト基盤を備え、**別プロジェクトの雛形（ボイラープレート）としても使える構成**にしている。

> サービスとしての内容（何を作っているか）は [docs/サービス概要.md](docs/サービス概要.md) を参照。
> 🚧 開発中（個人開発・学習目的）

## 全体構成

```
ブラウザ ── Next.js (App Router / :3000) ── Go API (:8081) ── PostgreSQL (:5434)
```

設計の詳細は各ディレクトリのREADMEへ:

- [api/README.md](api/README.md) — バックエンド設計（構成・規約・選定理由・R1ロードマップ）
- [web/README.md](web/README.md) — フロントエンド設計（ルートグループ・BFF・選定理由・R2/R3ロードマップ）

## 技術スタックと選定理由

| 領域             | 技術                                                   | なぜ選んだか                                                                             |
| ---------------- | ------------------------------------------------------ | ---------------------------------------------------------------------------------------- |
| バックエンド     | **Go 標準 net/http**                                   | Go 1.22+で主要なルーティングが標準化。フレームワークの隠蔽なしに仕組みを学べる・依存ゼロ |
| DBアクセス       | **database/sql + pgx**（ORMなし）                      | SQLとDB挙動を自分で制御・学習対象にする。型安全化はR1でsqlc検討                          |
| マイグレーション | **sql-migrate**                                        | up/down/status の履歴管理。psqlでDDL直接実行は禁止                                       |
| 認証             | **自前JWT（HS256）+ bcrypt + httpOnly Cookie**         | トークンをブラウザJSに触れさせない。alg強制・同一401など攻撃対策を自分の手で実装         |
| フロント         | **Next.js App Router + TypeScript**                    | RSC/Client境界の設計・BFF（Route Handler）を実践                                         |
| API通信          | **REST + 素のfetch**（axios/GraphQL不使用）            | Nextのfetch拡張（キャッシュ）に乗る。GraphQLは規模条件を満たさず不採用（判断ログはdocs） |
| スタイル         | **Tailwind v4 + shadcn/ui**                            | コピーイン方式＝コードを所有。デザイントークンで一括テーマ変更                           |
| テスト           | **go test（テーブル駆動）/ Vitest + Testing Library**  | 「ユーザーから見える振る舞い」でテストする方針                                           |
| CI               | **GitHub Actions**（パスフィルタ・lint/test/build）    | ローカルとCIで同一コマンド・同一設定ファイル                                             |
| 開発体験         | **air / golangci-lint / Prettier / Makefile / Docker** | ホットリロード・品質ゲート・コマンド集約・環境再現性                                     |

## ディレクトリ構成

```
api/            Go バックエンド（フラット構成 → api/README.md）
web/src/        Next.js（→ web/README.md）
  app/          ルーティング（(public)/(guest)/(authenticated) ＝アクセス制御境界）
  components/   UI部品（ui/ は shadcn 生成物）
  lib/          API アクセス層（fetch の唯一の置き場・server/client 分離）
migrations/     DBスキーマ（sql-migrate: ddl/*.sql）＋シード
docs/           設計書・学習ログ・判断ログ
```

## セットアップ

```bash
# 0. 開発ツールのインストール（初回のみ）
go install github.com/rubenv/sql-migrate/...@latest                  # マイグレーション
go install github.com/air-verse/air@latest                           # ホットリロード
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest  # linter
# ※ ~/go/bin に PATH を通しておくこと

# 1. 環境変数の準備（初回のみ）
cp api/.env.example api/.env         # 必須: DATABASE_URL / JWT_SECRET（任意: PORT / WEB_ORIGIN）
cp web/.env.example web/.env.local   # NEXT_PUBLIC_API_URL を設定

# 2. DB 起動＆スキーマ適用＆シード投入
make db-up
make migrate-up
make seed      # テストユーザー投入（任意・下表参照）

# 3. 開発サーバー起動（別ターミナルで各々）
make dev-api   # Go API（air ホットリロード・:8081）
make dev-web   # Next.js（:3000）
```

その他のコマンドは `make help` で一覧表示（テスト: `make test` / lint: `make lint` / ビルド: `make build`）。

### フルDocker起動（ローカルにGo/Nodeが無くても動く）

```bash
make docker-up   # db + api(air) + web(next dev) を一括起動（初回は依存DLで数分）
make docker-down # 停止
```

- ソースはマウント注入のため、ホストでの編集がそのままホットリロードされる
- ローカル直起動（make dev-api / dev-web）とはポートを取り合うため**同時には使えない**
- DBはどちらのモードでも同じコンテナ・同じデータを共有する

## テストユーザー（make seed で投入・ローカル専用）

| メールアドレス       | パスワード  | ロール |
| -------------------- | ----------- | ------ |
| company1@example.com | password123 | 企業   |
| company2@example.com | password123 | 企業   |
| talent1@example.com  | password123 | 人材   |
| talent2@example.com  | password123 | 人材   |

`make seed` は何度実行しても安全（upsert）。同じ email のユーザーが既にいる場合は上表の資格情報に上書きされる。

## テスト

```bash
# api（Go）
cd api && go test ./...        # 一括実行（-v で詳細・-run 'Test名' で絞り込み）

# web（Vitest + Testing Library）
cd web && npm run test         # 一括実行（CI と同じ）
cd web && npm run test:watch   # 変更を監視して自動再実行（開発時）
```

## Lint / Format

```bash
# api（Go）: golangci-lint（インストール: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest）
cd api && golangci-lint run ./...

# web（Next.js）: ESLint / Prettier
cd web && npm run lint          # コード品質（ESLint）
cd web && npm run format:check  # 整形チェック（Prettier）
cd web && npm run format        # 一括整形
```

## DBマイグレーション（sql-migrate）

スキーマ変更はすべて `migrations/ddl/*.sql` のマイグレーションで管理します（psql で直接 DDL を流さない）。

```bash
make migrate-up      # 未適用のマイグレーションをすべて適用
make migrate-status  # 適用状況の確認（APPLIED / PENDING）
make migrate-down    # ロールバック：直前の1件だけ巻き戻す
make -C migrations reset   # リセット：全件巻き戻し（全テーブルDROP。ローカル専用）
make migrate-new NAME=create_projects  # 新規マイグレーションの雛形作成
```

- `down` / `reset` は **DROP TABLE によりデータも消えます**。ローカル開発の手直し用で、本番では原則使いません（戻したいときは逆操作の Up を新規作成＝ロールフォワード）
- DBを完全に作り直したい場合: `docker compose down -v && docker compose up -d db && make migrate-up`（`-v` でデータボリュームごと削除）

## 🧬 このリポジトリを雛形として使う

### 使い方は2通り

**A. テンプレリポジトリ化（推奨・一度だけやればよい）**

1. このリポジトリを新しい名前でクローンし、下記の書き換えを実施
2. GitHub に push し、リポジトリの Settings → **Template repository** を ON
3. 以後の新プロジェクトは「**Use this template**」ボタン（または `gh repo create 新名前 --template 元リポジトリ`）で生成

**B. 直接クローンして作り変える**

```bash
git clone <このリポジトリ> my-app && cd my-app
rm -rf .git && git init   # 履歴を切り離す
```

### 書き換えポイント一覧

| #   | 対象                               | 現在の値                                                                              | 書き換え                                                                                                                                                                                                                                                             |
| --- | ---------------------------------- | ------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **プロジェクト識別子（全体一括）** | `tsunagu`                                                                             | 新プロジェクト名（小文字）。go.mod（tsunagu-api）・.golangci.yml（local-prefixes）・web/package.json（tsunagu-web）・docker-compose.yml（コンテナ名/DB名/ユーザー/パスワード/volume名）・migrations/dbconfig.yml・Makefile（psql の -U/-d）・api/.env.example に波及 |
| 2   | **表示名（全体一括）**             | `Tsunagu Works`                                                                       | 新サービス名。web の layout.tsx（metadata）・(guest)レイアウトのロゴ・各画面・not-found                                                                                                                                                                              |
| 3   | ポート（衝突時のみ）               | DB 5434 / API 8081 / web 3000                                                         | docker-compose.yml・api/.env（DATABASE_URL, PORT）・migrations/dbconfig.yml・web/.env.local                                                                                                                                                                          |
| 4   | **JWT_SECRET（必須）**             | —                                                                                     | `api/.env` に**新しいランダム値**を設定（テンプレ元の値を使い回さない）                                                                                                                                                                                              |
| 5   | デザイントークン                   | ブランド青 #1d4ed8 等                                                                 | `web/src/app/globals.css` の `:root`・`docs/デザインシステム.md`                                                                                                                                                                                                     |
| 6   | シードユーザー（任意）             | company1@example.com 等                                                               | `migrations/seed.sql`（メールアドレス・パスワードハッシュ）                                                                                                                                                                                                          |
| 7   | ドメイン固有ファイル（任意）       | 仕様ドラフト.html・docs/サービス概要.md・CLAUDE.md のドメイン記述・users の role 制約 | 新サービスに合わせて削除または書き換え                                                                                                                                                                                                                               |

### 手順（この順番で）

```bash
# 1. 一括置換（識別子 → 表示名の順。macOSのsedは -i '' が必要）
grep -rl 'tsunagu' --exclude-dir={.git,node_modules,.next} . | xargs sed -i '' 's/tsunagu/myapp/g'
grep -rl 'Tsunagu Works' --exclude-dir={.git,node_modules,.next} . | xargs sed -i '' 's/Tsunagu Works/My App/g'

# 2. Goモジュールの整合を確認
cd api && go mod tidy && cd ..

# 3. 環境変数を作り直す（JWT_SECRET は必ず新しい値に）
cp api/.env.example api/.env
cp web/.env.example web/.env.local

# 4. 動作確認チェックリスト（下記）
```

### 動作確認チェックリスト

- [ ] `grep -ri tsunagu . --exclude-dir={.git,node_modules,.next}` の残存がゼロ
- [ ] `make build`（api/web とも通る）
- [ ] `make test` / `make lint` が通る
- [ ] `make db-up && make migrate-up && make seed` が通る
- [ ] `make dev-api` + `make dev-web` → シードユーザーでログイン → 認証ガードが機能する

## ドキュメント

- [サービス概要](docs/サービス概要.md) — 何を作っているか（課題・信頼の設計・機能一覧）
- [アーキテクチャ設計書](docs/アーキテクチャ.md)
- [データ設計（ER図・リレーション）](docs/データ設計.md)
- [学習ログ](docs/学習ログ.md) — バックエンド / DB / フロントエンド / 開発環境
- 仕様デッキ: `仕様ドラフト.html`
