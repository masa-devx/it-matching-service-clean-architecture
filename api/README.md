# api — Go バックエンド設計（出発点）

Go 標準ライブラリ（net/http）によるREST API。tsunagu-works MVP の**フラット構成**（`package main` のファイル分割）を、リファクタの出発点としてそのまま保持している。

> ⚠️ このディレクトリは **Before（リファクタ対象）**。[後継リポジトリ設計プラン](../docs/後継リポジトリ設計プラン.md)に沿って、`apps/api-server/`（視点→層分割・スキーマ駆動・sqlc）へ段階的に作り替えていく。目標構成は末尾の[ロードマップ](#進化のロードマップ設計プランへ移行)へ。

## 構成とファイル責務

```
api/
├── main.go          # 組み立てだけ: 設定読込 → DB接続 → ルート登録 → サーバー起動/停止
├── config.go        # 環境変数の読み取りを一元化（envOr=任意+既定値 / mustEnv=必須+fail fast）
├── db.go            # 接続プール生成（上限明示・Ping疎通確認）
├── response.go      # writeJSON / writeError（エラー詳細はログのみ・クライアントには安全な文言）
├── middleware.go    # requestID / recover / logging / cors / requireAuth
├── logger.go        # 構造化ログ（slog）の初期化・レベル判定・相関ID付き子ロガー
├── jwt.go           # トークンの発行と検証（対で同居・HS256強制）
├── auth.go          # ドメイン機能: signup / login / me（機能ごとに1ファイルの型）
├── profile.go       # プロフィール（companies / talents）
├── projects.go      # 案件。ほか applications / contracts / work_reports /
│                    #   messages / masking / reviews が同じ型で並ぶ
├── *_status.go      # 状態遷移表（応募・掲載・契約・稼働報告）＝純粋関数＋テスト
├── health.go        # 死活確認（DB疎通込み）
├── .air.toml        # ホットリロード設定（開発ツール）
└── .golangci.yml    # linter設定（gosec / errorlint / exhaustive 等）
```

## リクエストの流れ

```
リクエスト → requestIDMiddleware（相関IDの発行・X-Request-ID）
           → recoverMiddleware（panic回復・スタックトレース記録）
           → loggingMiddleware（全記録・レベルはステータスで決定）
           → corsMiddleware（許可Origin判定・プリフライト応答）
           → mux（メソッド+パスでルーティング）
           → requireAuth（保護ルートのみ: Bearer検証 → userID を context へ）
           → handler（auth.go 等）→ writeJSON / writeError
```

## 設計の型（このリポジトリの規約）

- **main.go は組み立てだけ**。機能はドメイン語彙のファイルへ
- **SQL はプレースホルダ必須・`SELECT *` 禁止**（列を明示）
- **user_id は検証済みトークン（`userIDFrom(ctx)`）からのみ取得**（IDOR対策・入口一本化）
- **認証と認可を分ける**: `requireAuth`（誰か＝401）と `requireRole`（何をしてよいか＝403）を重ね掛けする
- **一覧で親子を出すときは JOIN 1回**（ループ内クエリ＝N+1 は禁止）。limit は上限をクランプする
- 重複チェックは SELECT 事前確認でなく **INSERT + 一意制約違反(23505)の検出**（TOCTOU対策）
- 認証失敗は**同一の401**（文言・ステータス・応答時間の3点で情報を漏らさない）
- エラーは `(結果, error)` + `fmt.Errorf("%w")` ラップ。`_ =` は「判断して捨てた」印（理由コメント付き）
- サーバーは**タイムアウト設定済み http.Server + graceful shutdown**（Slowloris対策・デプロイ時のリクエスト断絶防止）

## 技術選定と理由

| 選定                                 | 理由（詳細は docs/ の判断ログ）                                                                                                                                              |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **net/http（フレームワーク不使用）** | Go 1.22+ でメソッド指定・パスパラメータが標準化されルーティングの主要動機が消滅。仕組みを隠蔽せず学べる・依存ゼロ。後継設計では**ルーティング自体を oapi-codegen の生成コードに任せる** |
| **database/sql + pgx（ORM不使用）**  | SQL とDB挙動そのものが学習対象・状態機械のクエリを自分で制御。後継設計では **sqlc** を採用し、書いたSQLを資産として持ち越したまま型安全化する                                |
| **sql-migrate**                      | up/down/status の履歴管理。DDLは psql 直接実行禁止                                                                                                           |
| **golang-jwt/v5 + bcrypt**           | JWT はHS256を検証側で強制（algすり替え拒否）。パスワードは bcrypt（遅さが価値）                                                                                              |
| **air / golangci-lint**              | 開発マシンのツールとして go install（go.mod に入れない）                                                                                                                     |

## テスト

- `go test ./...`。テーブル駆動テスト（`auth_test.go` が見本）
- 検証ロジックは純粋関数に切り出してテスト可能に（`validateSignup`）

## 進化のロードマップ（設計プランへ移行）

旧計画（R1: handler / usecase / repository への分割）は、[後継リポジトリ設計プラン](../docs/後継リポジトリ設計プラン.md)に**置き換えられた**。repository 層は作らず、sqlc の生成する `Queries` をその代わりに使う。

```
現在: api/ フラット構成（handler内に検証・SQL・レスポンスが同居）
  ↓
目標: apps/api-server/internal/{company, talent, shared}
      handler（oapi-codegen 生成IFの実装）
        → usecase（業務ロジック・トランザクション境界）
        → generated/db（sqlc の Queries）        ※ repository 層は作らない
      shared/domain（状態遷移表・不変条件） ← DBにもHTTPにも依存しない
```

現行コードには目標構成への「切り込み線」が既に入っている：

| 現行コード | 移行先 |
| --- | --- |
| `*_status.go` の遷移表（純粋関数） | `shared/domain` |
| `validateSignup` などの検証関数 | `validator/` |
| `db.QueryRow` に渡している手書きSQL | `queries/*.sql`（sqlc の入力） |
| `requireAuth` / `requireRole` | `shared/auth` のミドルウェア（ロールはパスで一律） |
