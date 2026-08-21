# Tsunagu Works — アーキテクチャ設計版

**企業 × IT人材（副業・フリーランスエンジニア）のビジネスマッチング「Tsunagu Works」を、別の設計で再実装するリポジトリ。**

[tsunagu-works](https://github.com/masahiro96848/tsunagu-works)（フラット構成で MVP 完成済み）と**同じドメインを、スキーマ駆動・層分割・モノレポ構成で作り直す**。
2つのリポジトリを並べて「**どの規模で、どちらの設計を選ぶか**」という判断そのものを示すことがゴール。

> 設計の全体像: [docs/後継リポジトリ設計プラン.md](docs/後継リポジトリ設計プラン.md) ／ 個々の決定: [docs/adr/](docs/adr/README.md)
> `web/` は tsunagu-works MVP を土台に段階的に作り替えていく（[ADR-0006](docs/adr/0006-single-frontend-after-only.md)）。Before の姿は tsunagu-works リポジトリを参照。

## なぜ作り直すか

tsunagu-works は機能を出し切ることを優先し、**あえて層を分けずに MVP を完成させた**（その判断の記録は [docs/アーキテクチャ.md](docs/アーキテクチャ.md)）。
同じリポジトリでリファクタすると Before の姿が git 履歴の中にしか残らないため、**別リポジトリで再実装し、2つを並べて見せる**方針を採った。

ここでは、tsunagu-works で意図的にやらなかった設計に取り組む。

| 要素             | tsunagu-works（Before）      | このリポジトリ（After）               | 判断の理由                                                                       |
| ---------------- | ---------------------------- | ------------------------------------- | -------------------------------------------------------------------------------- |
| ルーティング     | `mux.HandleFunc` を手書き    | **生成コードが張る**（oapi-codegen）  | URL の一次情報を仕様に一本化                                                     |
| ハンドラの型     | `http.ResponseWriter` 直書き | **StrictHandler**（型付き req/res）   | 仕様とのズレがコンパイルエラーになる                                             |
| 依存の持ち方     | パッケージ変数の `db`        | **main.go で DI**                     | テストで接続を差し替えられる                                                     |
| DBアクセス       | `database/sql`（手書きSQL）  | **sqlc**                              | SQL を資産として残したまま型安全にする                                           |
| repository 層    | 無し                         | **無し（意図的）**                    | sqlc の `Queries` が repository 相当。薄皮を重ねない                             |
| 認可のロール判定 | 各ハンドラで `if user.Role`  | **パスプレフィックス × ミドルウェア** | 書き忘れが構造的に起きない                                                       |
| 一覧の取得       | 全件取得                     | **seek 法ページネーション**           | OFFSET は深いページで遅く、挿入でズレる                                          |
| テスト           | 純粋関数のみ（実装の25%）    | **実DBテスト（Tx分離）＋ factories**  | モックが多いと「テストは通るが動かない」が起きる                                 |
| フロントの分割   | 単体リポジトリ・手書き型     | **モノレポ（Turborepo）＋ 生成型**    | 1アプリ維持は規模判断（[ADR-0006](docs/adr/0006-single-frontend-after-only.md)） |
| API仕様          | 手書きの型（Go / TS で別々） | **TypeSpec → OpenAPI → Go/TS 生成**   | 1つの仕様から両方に同じ型が届く                                                  |
| 公開             | ローカルのみ                 | **Cloud Run ＋ Cloud Run Job**        | マイグレーションは1回だけ実行される必要がある                                    |

各判断の詳細と全対比は [設計プラン §15](docs/後継リポジトリ設計プラン.md) へ。

## 目標アーキテクチャ

```
<repo>/                        ★ Turborepo モノレポ
├─ api-server/                 # Go — internal/{company,talent,shared} の視点→層分割
│  ├─ generated/               #   oapi-codegen の生成物（編集禁止・コミットする）
│  └─ internal/company/handler #   StrictServerInterface の実装
├─ web/                        # Next.js 1アプリ（ロール分岐はルートグループ）
│                              #   Stage 1 構成 → features / external 構成へ段階移行
├─ packages/
│  ├─ spec/                    # TypeSpec（API契約の一次情報・apps のどちらの持ち物でもない）
│  └─ typescript-config/       # tsconfig の共有ベース
└─ migrations/ddl/             # sql-migrate
```

```
packages/spec/shared/models.tsp（1箇所で定義）
   ├──→ openapi-company.yaml ──→ Go: generated/api/company ＋ TS（orval・#8 で導入）
   └──→ openapi-talent.yaml  ──→ Go: generated/api/talent（Phase 2 で追加）
```

- **仕様変更 → 生成 → ビルドを Turborepo のタスクグラフで連鎖**させ、「生成し忘れ」を構造的に無くす
- 認可はパスで一律に決める: `/company/*` は企業ロール、`/talent/*` は人材ロールをミドルウェアが強制
- 依存方向は仕組みで守る（golangci-lint の depguard / ESLint の import 制約）

## 進め方（Phase 0 〜 7）

**まず Phase 0 で「案件の作成」1エンドポイントだけを貫通させ、コード生成の品質を確かめてから全体を確定した。**

| Phase | 内容                                                                        | ここで示すこと                           | 状態      |
| ----- | --------------------------------------------------------------------------- | ---------------------------------------- | --------- |
| **0** | 1エンドポイント貫通（TypeSpec → oapi-codegen / sqlc / orval → フォーム1つ） | 生成の質・Turborepo に Go を載せる現実性 | ✅ 完了   |
| 1     | 土台（層構成・DI・実DBテスト基盤・CI・依存方向の強制）                      | 設計を仕組みで守る                       | ✅ 完了   |
| 2     | 認証 ＋ プロフィール（company / talent の2系統）                            | 認可がパスで一律になる                   | ✅ 完了   |
| 3     | 案件（seek ページネーション・検索）                                         | 層分けした CRUD                          | ✅ 完了   |
| 4     | 応募（状態機械）                                                            | domain 層の見せ場                        | ✅ 完了   |
| 5     | デプロイ（Cloud Run・Cloud Run Job・CI/CD）                                 | 動くものを公開する                       | 🚧 進行中 |
| 6     | E2E・トレーシング                                                           | 通しで壊れないことを示す                 | —         |
| 7     | eKYC（multipart・Cloud Storage・署名付きURL）                               | 既存の土台に新種の機能を足す             | —         |

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

## 公開環境（本番）

**web**: https://tsunagu-web-985660768358.asia-northeast1.run.app ／ **api**: https://tsunagu-api-985660768358.asia-northeast1.run.app/health

```
main マージ ─→ GitHub Actions（WIF・キーレス認証）─→ Cloud Build（cloudbuild.yaml が一次情報）
                 build & push（タグ=コミットSHA）→ migrate Job --wait → Cloud Run 切り替え
                          │
             Artifact Registry ── Cloud Run（api / web）── Neon PostgreSQL
```

- **本番にシードデータは入れない**（動作確認は画面の signup から。下記テストアカウントはローカル専用）
- コールドスタート採用（min-instances 0）のため、初回アクセスは数秒かかることがある
- 手順・判断の詳細は [docs/デプロイ.md](docs/デプロイ.md)・[ADR-0011](docs/adr/0011-web-on-cloud-run.md)

## 動かす

```
ブラウザ ── Next.js web (:3001) ── Go api-server (:8082) ── PostgreSQL (:5435)
```

```bash
# 0. 前提ツール（初回のみ）: Node 24（.nvmrc）・pnpm（corepack）・Go 1.26
# ※ sql-migrate は go.mod の tool ディレクティブで管理（go install 不要・make 経由で自動解決）
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest  # linter（~/go/bin にPATHを通す）

# 1. 環境変数（初回のみ）
cp web/.env.example web/.env.local   # NEXT_PUBLIC_API_URL

# 2. 依存の取得と DB 起動・スキーマ適用
pnpm install
make db-up && make migrate-up

# 3. 開発用シードデータ（任意・冪等）
make seed

# 4. 開発サーバー（Go と Next.js が並列起動・Ctrl+C で両方停止）
pnpm dev
```

シード投入後のテストアカウント（パスワードは全員 `password123`・開発専用）:

| email                                       | ロール |
| ------------------------------------------- | ------ |
| company1@example.com / company2@example.com | 企業   |
| talent1@example.com / talent2@example.com   | 人材   |

- ビルド・検証は turbo に集約: `pnpm turbo build` ／ CI と同一の全チェックは `pnpm turbo lint format:check test build`
- Go テストの見やすい実行: `make test-api`（gotestsum・テスト名と PASS/FAIL を色付き表示）
- `make` は DB 専用の道具箱（`make help` で一覧）。マイグレーションは sql-migrate で管理（psql で DDL を直接流さない）

## ドキュメント

- **[後継リポジトリ設計プラン](docs/後継リポジトリ設計プラン.md)** — 設計の全体像（技術選定・認可設計・テスト戦略・Phase 計画）
- **[ADR](docs/adr/README.md)** — 個々の設計判断の記録（経緯・却下した代替案）
- [アーキテクチャ設計書](docs/アーキテクチャ.md) / [web/README.md](web/README.md) — 出発点（tsunagu-works MVP）の設計記録
- [サービス概要](docs/サービス概要.md) — 何を作っているか（課題・信頼の設計・機能一覧）
- [データ設計（ER図・リレーション）](docs/データ設計.md)
- [学習ログ](docs/学習ログ.md) — バックエンド / DB / フロントエンド / 開発環境
- 仕様デッキ: `仕様ドラフト.html`
