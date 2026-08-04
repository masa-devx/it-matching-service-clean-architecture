# api — Go バックエンド設計

Go 標準ライブラリ（net/http）によるREST API。**Stage 1: フラット構成**（`package main` のファイル分割）で運用中。

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
- 重複チェックは SELECT 事前確認でなく **INSERT + 一意制約違反(23505)の検出**（TOCTOU対策）
- 認証失敗は**同一の401**（文言・ステータス・応答時間の3点で情報を漏らさない）
- エラーは `(結果, error)` + `fmt.Errorf("%w")` ラップ。`_ =` は「判断して捨てた」印（理由コメント付き）
- サーバーは**タイムアウト設定済み http.Server + graceful shutdown**（Slowloris対策・デプロイ時のリクエスト断絶防止）

## 技術選定と理由

| 選定                                 | 理由（詳細は docs/ の判断ログ）                                                                                                                                              |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **net/http（フレームワーク不使用）** | Go 1.22+ でメソッド指定・パスパラメータが標準化されルーティングの主要動機が消滅。仕組みを隠蔽せず学べる・依存ゼロ。痛み（ミドルウェアの重ね掛け煩雑化等）が出たら chi を検討 |
| **database/sql + pgx（ORM不使用）**  | SQL とDB挙動そのものが学習対象・状態機械のクエリを自分で制御。スキャンが辛くなったら sqlx、R1 で repository 層を切るとき sqlc（型安全なコード生成）を検討                    |
| **sql-migrate**                      | up/down/status の履歴管理。参考実装と同型。DDLは psql 直接実行禁止                                                                                                           |
| **golang-jwt/v5 + bcrypt**           | JWT はHS256を検証側で強制（algすり替え拒否）。パスワードは bcrypt（遅さが価値）                                                                                              |
| **air / golangci-lint**              | 開発マシンのツールとして go install（go.mod に入れない）                                                                                                                     |

## テスト

- `go test ./...`。テーブル駆動テスト（`auth_test.go` が見本）
- 検証ロジックは純粋関数に切り出してテスト可能に（`validateSignup`）

## 進化のロードマップ（R1）

**発動条件**: ハンドラが太った・状態機械（Phase 3〜）のテストを書きたくなったとき。

```
現在: handler内に検証・SQL・レスポンスが同居
  ↓ R1
handler（HTTP入出力） → usecase（業務ロジック・状態遷移・テスト対象の本丸） → repository（SQL）
```

`validateSignup`（純粋関数）は usecase の種、`db.QueryRow` 部分は repository の種として、切り込み線を入れてある。
