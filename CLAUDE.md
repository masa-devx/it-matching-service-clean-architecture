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

詳細仕様: `仕様ドラフト.html` ／ Obsidian（コンセプト・仕様の正）: `~/Desktop/obsidian/20_projects/personal-apps/tsunagu-works/tsunagu-works.md`

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

## ドメイン仕様の核心（毎セッションの前提知識）

> マッチングの本質＝出会わせることではなく、**知らない者同士が安心して取引できる仕組み**。独自機能はすべて「信頼の設計」に集中させる。

**信頼の設計・4本柱**

- **エスクロー決済**: 仮払い→検収→支払い。「未払い」と「前払い」のリスクを同時に解決
- **連絡先マスキング**: メッセージ内のメール/電話を検出して伏せる（直接取引防止＝事業の生命線。原文は監査用に保存）
- **レビュー同時公開**: 両者提出まで非公開→同時公開（報復レビュー防止・Airbnb方式）
- **稼働報告**: 週次レポートで「働きぶりが見えない」不安と「証拠を残したい」を同時解決

**状態機械3本（設計の見せ場）**

```
① 選考: 応募 → 書類 → 面談 → 承諾(ダブルオプトイン) / 見送り / 取り下げ
② 契約: 成立 → 稼働中 → 検収待ち → 完了 / 中止（差し戻し→稼働中）
③ 決済: 仮払い待ち → 仮払い済み(エスクロー) → 検収OK → 支払い実行 → 完了 / 返金
※ ②業務と③お金を「別の状態機械に分離して同期」させるのが核心
```

**データ設計（コア）**: `users / companies / talents / projects / applications / contracts / work_reports / payments(idempotency_key) / messages(masked_body) / reviews(submitted_at・published_at) / notifications`

**MVPビルド順**: 環境(Phase0) → 認証＋プロフィール(1) → 案件CRUD＋検索(2) → 応募〜承諾＝状態機械①(3) → 契約＋稼働報告＋検収＝状態機械②(4) → マスキング＋レビュー同時公開(5)＝MVP完成 → 拡張: Stripeエスクロー・ダッシュボード・1万件計測

**スコープ外**: AI機能・リアルタイムチャット・モバイル・**KYC（本人確認）**（完成優先・2026-08-04にKYC除外を決定）

---

## 参照ドキュメント（必要時に読む・Obsidian側）

| ドキュメント | パス | 読むタイミング |
| --- | --- | --- |
| コンセプト・仕様の正 | `~/Desktop/obsidian/20_projects/personal-apps/tsunagu-works/tsunagu-works.md` | 仕様の疑問・決定ログ確認時 |
| 画面イメージ（仕様デッキ） | 同フォルダ `tsunagu-works-仕様デッキ.html` | 画面実装の着手時 |
| 参考実装の設計研究 | `~/Desktop/obsidian/02_notes/references/it-support-serviceの設計研究.md` | アーキテクチャ判断・模倣元の確認時 |
| 学習ロードマップ | `~/Desktop/obsidian/10_career/learning-roadmap/` | 学習の優先順位に迷ったとき |

---

## ディレクトリ構成（Stage 1：シンプル開始）

```
api/            # Go: フラット構成（main.go / db.go / auth.go / projects.go …機能ごと1ファイル）
migrations/     # DBスキーマ（sql-migrate: ddl/*.sql + dbconfig.yml + Makefile）
web/src/
  app/          # (public)/(guest)/(authenticated) ルートグループ＝アクセス制御の境界。page は薄く
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
- **ルートグループ `(public)` / `(guest)` / `(authenticated)` を崩さない**（アクセス制御の境界。ガードは各グループのlayoutに集約）。page.tsx は薄く保つ
- **api の main.go は「組み立て」だけ**。機能はドメイン語彙のファイル（projects.go 等）に分け、太ったら分割
- **user_id は必ず検証済みトークンから取る**（クライアント供給値を信用しない＝IDOR対策）
- SQL は**プレースホルダ必須**・`SELECT *` 禁止・状態遷移の定義はコード上で1か所にまとめる

詳細ルール:
@.claude/rules/workflow.md
@.claude/rules/backend.md
@.claude/rules/frontend.md

---

## セッション開始時の手順（現在地の把握・必ず最初に行う）

新しいセッションを開始したら、実装に入る前に以下で現在地を把握する:

1. `git branch --show-current` と `git status` で作業中のブランチ・未コミット変更を確認
2. `gh issue list --state open` で残タスクを確認（ブランチ名 `phase{N}/{Issue番号}-…` が現在作業中の Issue を指す）
3. 直近の学びの文脈が必要なら `docs/学習ログ.md` を参照
4. 仕様・コンセプトの確認が必要なら上記 Obsidian の仕様ノートを読む

把握した現在地（ブランチ・対応Issue・次のステップ）を最初に短く報告してから作業に入ること。

---

## 開発の進め方（理解して作る・厳守）

> 目的は「動くものを量産する」ことではなく、**全行を自分の言葉で説明できる状態で作る**こと。

1. **実装計画の提示 → 承認 →着手（必須）**: Issue着手時、コードを書く前に**実装計画**を提示し、ユーザーの承認を得る。計画に含めるもの：
   - 作成・変更するファイル一覧と各ファイルの役割
   - **コミット単位のステップ分割**（1ステップ＝1コミット目安。各ステップは単体でビルドが通る単位にする）
   - 使う技術・新出概念の予告
   - 各ステップの動作確認方法
   - **承認が出るまで実装しない**
2. **1ステップずつ実装し、ステップごとに止まる**: 実装 → ウォークスルー解説 → 動作確認の案内 → **ユーザーがコードリーディング・動作確認・コミットするのを待って**次のステップへ（Issue全体を一気に作らない）
3. 解説の型は固定：
   **何をした → 生成物の中身（ファイル/コードのウォークスルー） → なぜこう作るか → 新出の概念 → ✅実務のベストプラクティス → ⚠️アンチパターン → 理解度チェック□**
   - **ウォークスルー**＝生成・変更したファイルの中身をブロックごとに引用し、「この行/設定が何をしているか」をユーザーがコードを読みながら追える形で説明する（ただ結果を報告するだけにしない）
4. **理解度ゲート**: 理解度チェックに自分の言葉で答えられない項目があれば、質問→解消してから次へ（**分からないままマージしない**）
5. **コミットはユーザー自身がステップごとに細かく行う**（Claude は `git add` / `git commit` / `git push` をしない。`git status` / `diff` / `log` の確認は可）
6. **PR本文の「対応内容」はユーザーが自分の言葉で書く**（理解の最終確認。Claudeは下書きを求められたときだけ補助）
7. 各ステップの学びは `docs/`（バックエンド.md / DB.md / フロントエンド.md）に蓄積する
8. **サーバーの起動・停止はユーザーが行う**（Claude は build / vet / test までで検証）
9. 理解度チェック・「なぜ」の質問にはいつでも脱線して答える（学習が主目的）

## コードコメントの方針（公開リポジトリ品質）

- コード内コメントは**実務品質**：「なぜそうするか（意図・理由・注意）」だけを書く。学習用の逐行解説はコードに書かない
- 学習向けの詳細解説は**チャットと `docs/` 学習ログ側**に置く（リポジトリは採用担当が読む前提を保つ）

## 実装を渡す前の検証（Claude 必須）

- api を触ったら: `go vet ./...` && `go build ./...`
- web を触ったら: `npm run build`（型チェック込み）
- 上記が通った状態で解説し、ユーザーに引き渡す（コミット可能な状態にしておく）

## ブランチ・コミット・PR（ユーザーが実施）

- **Issue駆動開発**: 1 Issue ＝ 1ブランチ ＝ 1 PR
- **親子Issue**: 親＝機能のまとまり（Epic。マイルストーンに紐付け・機能全体の受け入れ条件を記載）／サブ＝作業単位。**ブランチ・PRを紐づけるのはサブIssueのみ**
- 親Issueは全サブ完了後、**ユーザーが機能として動作確認してから手動クローズ**（受け入れ確認）

### ブランチ命名規則

| 種類 | 形式 | 例 |
| --- | --- | --- |
| 機能実装（サブIssue） | `phase{N}/{Issue番号}-{内容}` | `phase0/3-signup` |
| リファクタ | `refactor/{内容}` | `refactor/api-clean-architecture` |
| ドキュメント | `docs/{内容}` | `docs/architecture` |
| バグ修正 | `fix/{Issue番号}-{内容}` | `fix/21-login-redirect` |

- **phase番号はマイルストーン番号と対応**（M0 → `phase0`、M1 → `phase1` …）
- 内容は英小文字ケバブケース・短く（2〜3語まで）
- **ブランチの作成・切り替えは Claude が行う**：ユーザーが **`/start <Issue番号>`**（または「#3へ」等の指示）で Issue を選択したら、Claude が main を最新化 → 該当ブランチ（`phase0/3-signup` 等）を作成・切り替えてから実装に入る。作業前にブランチが正しいか `git branch` で確認する
- **PR作成はユーザーが `/pr` を実行したときのみ** Claude が行う（テンプレ準拠・`Close #N` 自動付与。勝手に作らない）

### コミット・PR

- コミットは**機能ごとに小さく**。メッセージは日本語で「何を・なぜ」＋Issue番号（例: `feat: signup実装 (#3)`）
- PR本文に `Close #N` を書き、**ユーザーが確認してセルフマージ**（Claudeはマージしない）
- テンプレート: `.github/ISSUE_TEMPLATE/`（feature / bug）・`.github/pull_request_template.md`

## 用語集（ユビキタス言語・コード命名はこれに従う）

| 日本語 | コード上の語彙 | 意味 |
| --- | --- | --- |
| 企業 | `company` | 発注側ユーザー |
| 人材 | `talent` | 受注側ユーザー（副業・フリーランスエンジニア） |
| 案件 | `project` | 企業が掲載する仕事 |
| 応募 | `application` | 人材→案件への応募（状態機械①） |
| 契約 | `contract` | 承諾後の取引単位（状態機械②） |
| 稼働報告 | `work_report` | 週次の作業レポート |
| 検収 | `acceptance` | 企業による成果確認 |
| レビュー | `review` | 相互評価（同時公開） |

---

## 環境変数

- 実値は `.env`（gitignore）、キー名の見本は `.env.example` をコミット
- **必須**（未設定なら起動失敗）: `DATABASE_URL` / `JWT_SECRET`
- **任意**（安全なデフォルトあり）: `PORT`（既定8081） / `WEB_ORIGIN`（既定 http://localhost:3000） / `NEXT_PUBLIC_API_URL`
- api側の環境変数の読み取りは `api/config.go` に集約（ハードコード禁止）
