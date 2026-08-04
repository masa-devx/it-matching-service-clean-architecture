# Tsunagu Works

**企業 × IT人材（副業・フリーランスエンジニア）のビジネスマッチングサービスを模した開発**

「正社員採用は難しい、でも技術力が必要」な企業と、「副業を始めたい」エンジニアを、
**安心して取引できる仕組み（エスクロー決済・連絡先マスキング・相互レビューの同時公開）** つきでつなぎます。

> 🚧 開発中（個人開発）。Go + Next.js によるフルスタック構成。

---

## 解決する課題

マッチングサービスの価値は「出会わせること」ではなく、**知らない者同士が安心して取引できる仕組み**にあると考えています。

| 立場             | 課題                                                                 | 本サービスの解決                                    |
| ---------------- | -------------------------------------------------------------------- | --------------------------------------------------- |
| 企業（発注側）   | 誰が本当にできる人か見極められない／支払い・契約事務が面倒           | プロフィール・実績の構造化、KYC、ステータスの可視化 |
| IT人材（受注側） | 案件の入口がない／**納品したのに支払われないリスク**／実績が残らない | エスクロー決済、稼働報告、実績の蓄積                |
| プラットフォーム | 直接取引（中抜け）／報復レビュー                                     | 連絡先マスキング、レビューの同時公開                |

## 特徴（信頼の設計）

- **エスクロー決済** — 契約時に企業が仮払い → 検収 → 人材へ支払い。「未払い」と「前払い」のリスクを同時に解決する、お金の状態機械
- **連絡先マスキング** — メッセージ内のメールアドレス・電話番号を検出して伏せ、プラットフォーム外への誘導（直接取引）を抑止
- **レビュー同時公開** — 双方が書き終えるまで互いに非公開 → 同時公開。報復レビューを防ぎ評価の信頼性を守る
- **稼働報告** — 週次の作業レポートで「リモートで働きぶりが見えない」不安と「働いた証拠を残したい」ニーズを同時に解決

## 主な機能と実装状況

| 機能                                             | 状況      |
| ------------------------------------------------ | --------- |
| 認証（signup / login・JWT + httpOnly Cookie）    | 🔨 実装中 |
| 企業・人材プロフィール／KYC（本人確認）          | ⬜ 予定   |
| 案件の掲載・検索（スキル×単価×稼働のフィルタ）   | ⬜ 予定   |
| 応募 → 選考 → 承諾（ダブルオプトイン・状態遷移） | ⬜ 予定   |
| 契約・稼働報告・検収（状態遷移）                 | ⬜ 予定   |
| メッセージ（連絡先マスキング）                   | ⬜ 予定   |
| レビュー（同時公開）                             | ⬜ 予定   |
| エスクロー決済（Stripe）                         | ⬜ 予定   |

## 技術スタック

| 領域           | 技術                                                         |
| -------------- | ------------------------------------------------------------ |
| フロントエンド | Next.js（App Router）/ TypeScript / Tailwind CSS + shadcn/ui |
| バックエンド   | Go（net/http）/ JWT（bcrypt・HS256）                         |
| DB             | PostgreSQL 16（Docker）                                      |
| テスト         | go test / Vitest + Testing Library / Playwright（段階導入）  |

## アーキテクチャ

```
ブラウザ ── Next.js (App Router / :3000) ── Go API (:8081) ── PostgreSQL (:5434)
```

- **トークンをクライアント JS に触れさせない**：JWT は httpOnly Cookie に保存し、Next.js のサーバーサイドで読み取って Go API へ Bearer 転送
- **I/O の集約**：フロントの API 呼び出しは `web/src/lib/` に集約。バックエンドの共通処理はヘルパー／ミドルウェアに集約
- **認可の二重チェック**：ユーザーIDは検証済みトークンからのみ取得し、リソース所有は SQL レベルで検証
- **状態遷移の一元管理**：応募・契約・決済の各ステータスは遷移表で定義

詳細は [docs/アーキテクチャ.md](docs/アーキテクチャ.md) を参照。

## ディレクトリ構成

```
api/            Go バックエンド（ドメイン単位のファイル構成）
migrations/     DB スキーマ（sql-migrate: ddl/*.sql・up/down/status 管理）
web/src/
  app/          ルーティング（(public) / (guest) / (authenticated) ルートグループ）
  components/   UI 部品
  hooks/        カスタムフック
  lib/          API アクセス層（fetch の唯一の置き場・server/client 分離）
docs/           設計書・開発ドキュメント
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

## テストユーザー（make seed で投入・ローカル専用）

| メールアドレス | パスワード | ロール |
| --- | --- | --- |
| company1@example.com | password123 | 企業 |
| company2@example.com | password123 | 企業 |
| talent1@example.com | password123 | 人材 |
| talent2@example.com | password123 | 人材 |

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
make -C migrations up      # 未適用のマイグレーションをすべて適用
make -C migrations status  # 適用状況の確認（APPLIED / PENDING）
make -C migrations down    # ロールバック：直前の1件だけ巻き戻す
make -C migrations reset   # リセット：全件巻き戻し（全テーブルDROP。ローカル専用）
make -C migrations new NAME=create_projects  # 新規マイグレーションの雛形作成
```

- `down` / `reset` は **DROP TABLE によりデータも消えます**。ローカル開発の手直し用で、本番では原則使いません（戻したいときは逆操作の Up を新規作成＝ロールフォワード）
- DBを完全に作り直したい場合: `docker compose down -v && docker compose up -d && make -C migrations up`（`-v` でデータボリュームごと削除）

## ドキュメント

- [アーキテクチャ設計書](docs/アーキテクチャ.md)
- 仕様デッキ: `仕様ドラフト.html`
