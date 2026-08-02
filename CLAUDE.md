# CLAUDE.md — Tsunagu Works

必ず日本語で回答してください。

---

## セキュリティルール

- **`.env*` ファイル（`.env.example` を除く）を絶対に読み込まないこと**
  - `Read` / `cat` / `head` / `grep` 等で `.env` 系の中身を参照してはならない
  - 環境変数のキー名を確認したい場合は `.env.example` のみ参照すること
- **シークレット（JWT_SECRET・DBパスワード等）をコード・コミットメッセージに含めないこと**

---

## プロジェクト概要

**企業 × IT人材（副業・フリーランスエンジニア）のビジネスマッチング「Tsunagu Works」**。
[it-support-service](https://github.com/yamao-sys/it-support-service) を参考にした転職ポートフォリオ旗艦。独自機能＝**信頼の設計**（エスクロー決済・連絡先マスキング・レビュー同時公開・稼働報告）。

詳細仕様: `仕様ドラフト.html` ／ Obsidian: `20_projects/personal-apps/tsunagu-works/`

| 項目 | 内容 |
| --- | --- |
| バックエンド | Go（net/http・まずフラット構成 → R1でクリーンアーキへ） |
| フロント | Next.js App Router + TypeScript |
| 状態管理 | まず素のfetch → 後段で TanStack Query 導入 |
| スタイル | Tailwind CSS + shadcn/ui（導入は画面着手時） |
| DB | PostgreSQL（Docker） |
| 認証 | 自前JWT（bcrypt・httpOnly Cookie）＝ todo-app の型 |
| テスト | go test / Vitest / Playwright（段階導入） |

### ポート（他プロジェクトと衝突しない値に固定）

| 用途 | ポート | 備考 |
| --- | --- | --- |
| PostgreSQL | **5434** | 5432=ローカルPostgres、5433=todo-app が使用中 |
| Go API | **8081** | 8080=todo-app api と衝突回避 |
| Next.js | 3000 | |

---

## ディレクトリ構成（Stage 1：シンプル開始）

```
api/            # Go: フラット構成（main.go / db.go / auth.go / projects.go …機能ごと1ファイル）
web/src/
  app/          # (auth)/(main) ルートグループ。page は薄く
  components/   # UI部品
  hooks/        # カスタムフック
  lib/          # ★API呼び出しは必ずここに集約（コンポーネントに fetch を書かない）
docs/           # 学習ログ（学習ログ.md=目次 / バックエンド.md / DB.md / フロントエンド.md）
```

### 成長ロードマップ（痛みを感じたらリファクタ）

| 段階 | リファクタ | 発動条件 |
| --- | --- | --- |
| R1 | api: フラット → クリーンアーキ（handler/usecase/repository。`_templates/go-api-boilerplate` の形へ） | ハンドラが太った・状態機械のテストを書きたい |
| R2 | web: components/hooks/lib → `features/` ドメイン分割 | ドメインが3つ超で探すのが辛い |
| R3 | web: lib → `external/`（dto/handler/repository・server-only）＋ prefetch/Hydration | Zod検証・Server-first を入れたい |

> ゴール構成の参考: [next-app-router-architecture](https://github.com/YukiOnishi1129/next-app-router-architecture)（フロント）／ go-api-boilerplate（バック）

---

## アーキテクチャ原則（Stage 1 版）

- **fetch は `web/src/lib/` のみ**。コンポーネント・フックに直接書かない（将来の external 化の種）
- **ルートグループ `(auth)` / `(main)` を崩さない**。page.tsx は薄く保つ
- **api の main.go は「組み立て」だけ**。機能はドメイン語彙のファイル（projects.go 等）に分け、太ったら分割
- **user_id は必ず検証済みトークンから取る**（クライアント供給値を信用しない＝IDOR対策）
- SQL は**プレースホルダ必須**・`SELECT *` 禁止・状態遷移の定義はコード上で1か所にまとめる

詳細ルール:
@.claude/rules/workflow.md
@.claude/rules/backend.md
@.claude/rules/frontend.md

---

## 開発の進め方（todo-app 方式・厳守）

1. **小さく1ステップずつ**実装する（一気に作らない）
2. 実装したら**必ず詳細解説**（何を・なぜ・実務のベストプラクティス・アンチパターン）→ **ユーザーの確認/質問を待つ**
3. **コミットはユーザー自身が機能ごとに小さく行う**（Claude は `git add` / `git commit` / `git push` をしない。`git status` / `diff` / `log` の確認は可）
4. 各ステップの学びは `docs/`（バックエンド.md / DB.md / フロントエンド.md）に蓄積する
5. **サーバーの起動・停止はユーザーが行う**（Claude は build / vet / test までで検証）
6. 理解度チェック・「なぜ」の質問にはいつでも脱線して答える（学習が主目的）

## 実装を渡す前の検証（Claude 必須）

- api を触ったら: `go vet ./...` && `go build ./...`
- web を触ったら: `npm run build`（型チェック込み）
- 上記が通った状態で解説し、ユーザーに引き渡す（コミット可能な状態にしておく）

## ブランチ・コミット（ユーザーが実施）

- ブランチ: `main` ＋ 作業ブランチ（`feat/xxx` / `fix/xxx` / `refactor/xxx` / `docs/xxx`）
- 大きな変更（機能追加・リファクタR1〜R3）はブランチを切り、完了後 main にマージ
- コミットは**機能ごとに小さく**。メッセージは日本語で「何を・なぜ」

---

## 環境変数

- 実値は `.env`（gitignore）、キー名の見本は `.env.example` をコミット
- `DATABASE_URL` / `JWT_SECRET` / `NEXT_PUBLIC_API_URL` を想定
