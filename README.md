# Tsunagu Works — アーキテクチャ設計版

**企業 × IT人材（副業・フリーランスエンジニア）のビジネスマッチング「Tsunagu Works」を、別の設計で再実装するリポジトリ。**

[tsunagu-works](https://github.com/masahiro96848/tsunagu-works)（フラット構成で MVP 完成済み）と**同じドメインを、スキーマ駆動・層分割・モノレポ構成で作り直す**。
2つのリポジトリを並べて「**どの規模で、どちらの設計を選ぶか**」という判断そのものを示すことがゴール。

> 設計の全体像と判断の記録: [docs/後継リポジトリ設計プラン.md](docs/後継リポジトリ設計プラン.md)
> 現在のコードは tsunagu-works MVP のスナップショット（＝リファクタの出発点）。ここから段階的に作り替えていく。

## なぜ作り直すか

tsunagu-works は機能を出し切ることを優先し、**あえて層を分けずに MVP を完成させた**（その判断の記録は [docs/アーキテクチャ.md](docs/アーキテクチャ.md)）。
同じリポジトリでリファクタすると Before の姿が git 履歴の中にしか残らないため、**別リポジトリで再実装し、2つを並べて見せる**方針を採った。

ここでは、tsunagu-works で意図的にやらなかった設計に取り組む。

| 要素             | tsunagu-works（Before）      | このリポジトリ（After）                     | 判断の理由                                           |
| ---------------- | ---------------------------- | ------------------------------------------- | ---------------------------------------------------- |
| ルーティング     | `mux.HandleFunc` を手書き    | **生成コードが張る**（oapi-codegen）        | URL の一次情報を仕様に一本化                         |
| ハンドラの型     | `http.ResponseWriter` 直書き | **StrictHandler**（型付き req/res）         | 仕様とのズレがコンパイルエラーになる                 |
| 依存の持ち方     | パッケージ変数の `db`        | **main.go で DI**                           | テストで接続を差し替えられる                         |
| DBアクセス       | `database/sql`（手書きSQL）  | **sqlc**                                    | SQL を資産として残したまま型安全にする               |
| repository 層    | 無し                         | **無し（意図的）**                          | sqlc の `Queries` が repository 相当。薄皮を重ねない |
| 認可のロール判定 | 各ハンドラで `if user.Role`  | **パスプレフィックス × ミドルウェア**       | 書き忘れが構造的に起きない                           |
| 一覧の取得       | 全件取得                     | **seek 法ページネーション**                 | OFFSET は深いページで遅く、挿入でズレる              |
| テスト           | 純粋関数のみ（実装の25%）    | **go-txdb ＋ factories（実DB）**            | モックが多いと「テストは通るが動かない」が起きる     |
| フロントの分割   | 1アプリ内でロール分岐        | **company / talent の2アプリ ＋ Turborepo** | ロール分岐がアプリ境界になる                         |
| API仕様          | 手書きの型（Go / TS で別々） | **TypeSpec → OpenAPI → Go/TS 生成**         | 1つの仕様から両方に同じ型が届く                      |
| 公開             | ローカルのみ                 | **Cloud Run ＋ Cloud Run Job**              | マイグレーションは1回だけ実行される必要がある        |

各判断の詳細と全対比は [設計プラン §15](docs/後継リポジトリ設計プラン.md) へ。

## 目標アーキテクチャ

```
<repo>/                        ★ Turborepo モノレポ（目標構成）
├─ apps/
│  ├─ api-server/              # Go — internal/{company,talent,shared} の視点→層分割
│  ├─ company/                 # Next.js — 企業向け
│  └─ talent/                  # Next.js — 人材向け
├─ packages/
│  ├─ spec/                    # TypeSpec（API契約の一次情報）
│  ├─ api-client/              # orval 生成物（型 + Fetch Client + Zod）
│  └─ ui/                      # 2アプリで共有するデザインシステム
└─ migrations/ddl/             # sql-migrate
```

```
packages/spec/shared/models.tsp（1箇所で定義）
   ├──→ openapi-company.yaml ──→ Go: gen/api/company ＋ TS: api-client/company
   └──→ openapi-talent.yaml  ──→ Go: gen/api/talent  ＋ TS: api-client/talent
```

- **仕様変更 → 生成 → ビルドを Turborepo のタスクグラフで連鎖**させ、「生成し忘れ」を構造的に無くす
- 認可はパスで一律に決める: `/company/*` は企業ロール、`/talent/*` は人材ロールをミドルウェアが強制
- 依存方向は仕組みで守る（golangci-lint の depguard / ESLint の import 制約）

## 進め方（Phase 0 〜 7）

**まず Phase 0 で「案件の作成」1エンドポイントだけを貫通させ、コード生成の品質を確かめてから全体を確定する。**

| Phase | 内容                                                                        | ここで示すこと                           | 状態        |
| ----- | --------------------------------------------------------------------------- | ---------------------------------------- | ----------- |
| **0** | 1エンドポイント貫通（TypeSpec → oapi-codegen / sqlc / orval → フォーム1つ） | 生成の質・Turborepo に Go を載せる現実性 | 🔜 次はここ |
| 1     | 土台（層構成・DI・go-txdb・CI・依存方向の強制）                             | 設計を仕組みで守る                       | —           |
| 2     | 認証 ＋ プロフィール（company / talent の2系統）                            | 認可がパスで一律になる                   | —           |
| 3     | 案件（seek ページネーション・検索）                                         | 層分けした CRUD                          | —           |
| 4     | 応募（状態機械）                                                            | domain 層の見せ場                        | —           |
| 5     | デプロイ（Cloud Run・Cloud Run Job・CI/CD）                                 | 動くものを公開する                       | —           |
| 6     | E2E・トレーシング                                                           | 通しで壊れないことを示す                 | —           |
| 7     | eKYC（multipart・Cloud Storage・署名付きURL）                               | 既存の土台に新種の機能を足す             | —           |

**実装スコープは4ドメインのみ**（認証＋プロフィール / 案件 / 応募 / eKYC）。
契約・稼働報告・メッセージ・レビュー・エスクローは移植しない——**同じパターンの繰り返しで、新しく示せることが少ない**ため。機能の量は tsunagu-works が担い、このリポジトリは**設計を示す**ことに集中する。

## ドメイン: 何のサービスか

マッチングの本質は「出会わせること」ではなく、**知らない者同士が安心して取引できる仕組み**をつくること。Tsunagu Works はその「信頼の設計」に独自機能を集中させている。

| 不安                                     | 機能                                                   |
| ---------------------------------------- | ------------------------------------------------------ |
| 「プラットフォーム外で取引されてしまう」 | **連絡先マスキング**（メール・電話・URLを伏せる）      |
| 「悪い評価をつけたら報復されそう」       | **レビュー同時公開**（両者提出まで非公開）             |
| 「働きぶりが見えない / 証拠を残したい」  | **週次の稼働報告**（提出→承認/差し戻し）               |
| 「報酬が支払われないかもしれない」       | **エスクロー決済**（構想。両リポジトリともスコープ外） |

これらは tsunagu-works で実装済み。ドメインの詳細は [docs/サービス概要.md](docs/サービス概要.md) と [docs/データ設計.md](docs/データ設計.md) へ。

**出発点の規模**（tsunagu-works MVP・2026-08-07 時点）: テーブル9・APIエンドポイント28・画面23・Go実装3,948行（テスト1,309行）・状態遷移表5本。

## 現在のコードを動かす（出発点＝MVP構成）

> ⚠️ 以下は**リファクタ前の現行構成**（Go フラット構成 + Next.js 1アプリ）の手順。Phase 0 以降、Turborepo 構成へ段階的に置き換わる。

```
ブラウザ ── Next.js (App Router / :3001) ── Go API (:8082) ── PostgreSQL (:5435)
```

```bash
# 0. 開発ツール（初回のみ・~/go/bin に PATH を通す）
go install github.com/rubenv/sql-migrate/...@latest
go install github.com/air-verse/air@latest
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# 1. 環境変数（初回のみ）
cp api/.env.example api/.env         # 必須: DATABASE_URL / JWT_SECRET
cp web/.env.example web/.env.local   # NEXT_PUBLIC_API_URL

# 2. DB 起動・スキーマ適用・シード投入
make db-up && make migrate-up && make seed

# 3. 開発サーバー（別ターミナルで各々）
make dev-api   # Go API（air ホットリロード・:8082）
make dev-web   # Next.js（:3001）
```

- その他のコマンドは `make help`（テスト: `make test` / lint: `make lint` / ビルド: `make build`）
- ローカルに Go / Node が無い場合は `make docker-up` で一括起動（`make docker-down` で停止）
- マイグレーションは sql-migrate で管理（`make migrate-up` / `migrate-status` / `migrate-down`。psql で DDL を直接流さない）

### テストユーザー（`make seed` で投入・ローカル専用）

| メールアドレス                              | パスワード  | ロール |
| ------------------------------------------- | ----------- | ------ |
| company1@example.com / company2@example.com | password123 | 企業   |
| talent1@example.com / talent2@example.com   | password123 | 人材   |

性能計測用の大量データは `make seed-large`（案件5万件・決定的な連番データ。計測記録は [docs/性能.md](docs/性能.md)）。

## ドキュメント

- **[後継リポジトリ設計プラン](docs/後継リポジトリ設計プラン.md)** — このリポジトリの設計の全体像（技術選定・認可設計・テスト戦略・Phase 計画）
- [アーキテクチャ設計書](docs/アーキテクチャ.md) — 出発点（現行コード）の構成と「層を分けなかった判断」の記録
- [api/README.md](api/README.md) / [web/README.md](web/README.md) — 出発点の各実装の設計
- [サービス概要](docs/サービス概要.md) — 何を作っているか（課題・信頼の設計・機能一覧）
- [データ設計（ER図・リレーション）](docs/データ設計.md)
- [学習ログ](docs/学習ログ.md) — バックエンド / DB / フロントエンド / 開発環境
- 仕様デッキ: `仕様ドラフト.html`
