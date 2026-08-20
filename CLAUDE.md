# CLAUDE.md — Tsunagu Works アーキテクチャ再実装

必ず日本語で回答してください。

---

## セキュリティルール

- **`.env*` ファイル（`.env.example` を除く）を絶対に読み込まないこと**
  - `Read` / `cat` / `head` / `grep` 等で `.env` 系の中身を参照してはならない
  - 環境変数のキー名を確認したい場合は `.env.example` のみ参照すること
- **シークレット（JWT_SECRET・DBパスワード等）をコード・コミットメッセージに含めないこと**

---

## プロジェクト概要

**企業 × IT人材（副業・フリーランスエンジニア）のビジネスマッチング「Tsunagu Works」を、別の設計で再実装するリポジトリ。**

[tsunagu-works](https://github.com/masahiro96848/tsunagu-works)（フラット構成・MVP完成済み）と同じドメインを、スキーマ駆動・層分割・モノレポ構成で作り直す。**2つのリポジトリを並べて「どの規模で、どちらの設計を選ぶか」という判断を示す**のがゴール。MVPフェーズは終了しており、以後は通常のIssue駆動開発で進める。

- **設計の正**: `docs/後継リポジトリ設計プラン.md`（技術選定・認可設計・テスト戦略・Phase計画。疑問が出たらまずここ）
- **このリポジトリは After 専用**（[ADR-0006](docs/adr/0006-single-frontend-after-only.md)）。`web/` は tsunagu-works MVP を土台に作り替えていく。Before の参照は tsunagu-works リポジトリ

| 項目 | 内容 |
| --- | --- |
| モノレポ | Turborepo + pnpm workspace（Goもタスクグラフに載せ、仕様変更→生成→ビルドを連鎖させる） |
| API仕様 | TypeSpec → OpenAPI → oapi-codegen（Go・StrictHandler）/ orval（TS: 型+Fetch Client+Zod）。生成物はコミットし、CIで差分チェック |
| バックエンド | Go。`internal/{company,talent,shared}` の**視点→層**分割（handler / usecase / validator + shared/domain） |
| フロント | Next.js App Router + TypeScript。**web/ 1アプリ**（ロール分岐はルートグループ・ADR-0006） |
| DB | PostgreSQL（Docker）＋ **sqlc**（repository層は作らない。`Queries` が repository 相当） |
| サーバー状態 | TanStack Query（prefetch / Hydration）。書き込みは Server Actions |
| 認証・認可 | 自前JWT（bcrypt・httpOnly Cookie）。**ロール認可はパスプレフィックス（/company/*, /talent/*）×ミドルウェアで一律** |
| テスト | 実DBテスト（pgx.Tx分離・ADR-0008）+ factories・モックしない / API統合テスト（実MW・ADR-0012）/ Vitest |
| デプロイ | Cloud Run + Cloud Run Job（migrate）+ Neon |

### ポート（他プロジェクトと衝突しない値に固定）

| 用途 | ポート | 備考 |
| --- | --- | --- |
| PostgreSQL | **5435** | 5432=ローカルPostgres、5433=todo-app、5434=tsunagu-works |
| Go API | **8082** | 8080=todo-app、8081=tsunagu-works |
| Next.js（web） | **3001** | 3000=tsunagu-works web と衝突回避 |

（経緯の記録: [ADR-0005](docs/adr/0005-separate-ports-containers.md)）

---

## ドメイン仕様の核心（毎セッションの前提知識）

> マッチングの本質＝出会わせることではなく、**知らない者同士が安心して取引できる仕組み**。独自機能は「信頼の設計」（連絡先マスキング・レビュー同時公開・稼働報告・エスクロー構想）に集中させる。ドメインの詳細は `docs/サービス概要.md`。

### このリポジトリの実装スコープ（4ドメインのみ）

| ドメイン | 見せ場 |
| --- | --- |
| 認証 ＋ プロフィール | JWT・Cookie・ロールミドルウェア（company / talent の2系統） |
| 案件（CRUD・検索） | 層分けした CRUD・**seek ページネーション** |
| 応募（状態機械） | **遷移表を `shared/domain` に置く**＝domain 層の見せ場 |
| eKYC（ファイルアップロード） | multipart・Cloud Storage・署名付きURL・機微情報（Phase 7・talent側のみ） |

- **スコープ外**: 契約・稼働報告・メッセージ・レビュー・エスクロー決済（同じパターンの繰り返しで新しく示せることが少ない。機能の量は tsunagu-works が担う）、AI機能・リアルタイムチャット・モバイル
- ⚠️ tsunagu-works では KYC を除外したが、**このリポジトリでは eKYC はスコープ内**（ファイル操作・オブジェクトストレージの学習題材として Phase 7 で実装）
- **データ設計（コア・5テーブル）**: `users / companies / talents / projects / applications`

### Phase 0〜7（ブランチ命名の phase{N} はこの番号に対応）

```
0: 1エンドポイント貫通（生成品質の検証。ここで詰まったら設計見直し）← まずここだけやる
1: 土台（層構成・DI・実DBテスト基盤・CI・依存方向の強制）
2: 認証＋プロフィール（2系統・ロールMW） → 3: 案件 → 4: 応募（状態機械）
5: デプロイ（Cloud Run） → 6: E2E・トレーシング → 7: eKYC
```

---

## 参照ドキュメント（必要時に読む）

| ドキュメント | パス | 読むタイミング |
| --- | --- | --- |
| **後継リポジトリ設計プラン** | `docs/後継リポジトリ設計プラン.md` | **設計の疑問・Phase着手時（一次情報）** |
| 出発点の設計記録 | `docs/アーキテクチャ.md`・`web/README.md` | web の既存コードを読む・触るとき |
| コンセプト・仕様の正 | `~/Desktop/obsidian/20_projects/personal-apps/tsunagu-works/tsunagu-works.md` | 仕様の疑問・決定ログ確認時 |
| 画面イメージ（仕様デッキ） | 同フォルダ `tsunagu-works-仕様デッキ.html` | 画面実装の着手時 |
| 参考実装の設計研究 | `~/Desktop/obsidian/02_notes/references/` 配下 | アーキテクチャ判断で外部実装を確認したいとき |

---

## ディレクトリ構成

```
api-server/     Go（新構成）。internal/{company,talent,shared} の視点→層分割へ育てる
  generated/    oapi-codegen の生成物（編集禁止・コミットする）
web/            Next.js 1アプリ。Stage 1 構成 → features / external 構成へ段階移行
packages/       spec（TypeSpec・契約の一次情報）/ typescript-config
migrations/     sql-migrate（DDL・シード）
docs/           設計書・学習ログ・adr/（設計判断の記録）
```

- web の移行対応表（現行コードのどこが目標のどの層になるか）は `web/README.md` 末尾のロードマップを参照

---

## アーキテクチャ原則

**不変の原則（どちらの構成でも厳守）**

- **user_id / company_id は必ず検証済みトークンから取る**（クライアント供給値を信用しない＝IDOR対策）
- SQL は**プレースホルダ必須**・`SELECT *` 禁止（列を明示）
- **状態遷移の定義はコード上の1か所（遷移表・ホワイトリスト）**にまとめる
- 所有チェックは **SQL の WHERE に埋め込む**（取得してから判定しない）
- **未公開・原文は「取得しない」**（画面で隠さない）
- 確定した合意は**値でコピーする**（スナップショット）

**web/ の既存コード（Stage 1 構成）を触る場合**: fetch は `web/src/lib/` のみ・ルートグループ境界を崩さない（詳細は `.claude/rules/`）

**新構成（api-server/ packages/・web の作り替え部分）**: 設計プラン §4〜7 と ADR を正とする。依存方向は仕組みで強制する（golangci-lint の depguard / ESLint の import 制約）

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
4. 設計の疑問は `docs/後継リポジトリ設計プラン.md` を読む

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
- **生成物（`generated/` 等）は編集しない**。直したければ仕様（.tsp / queries）を直して再生成する

## 実装を渡す前の検証（Claude 必須）

- api を触ったら: `go vet ./...` && `go build ./...`
- web / packages を触ったら: `pnpm turbo build`（生成→ビルドの連鎖・型チェック込み）
- CI と同一の全チェック: `pnpm turbo lint format:check test build`
- 上記が通った状態で解説し、ユーザーに引き渡す（コミット可能な状態にしておく）

## ブランチ・コミット・PR（ユーザーが実施）

- **Issue駆動開発**: 1 Issue ＝ 1ブランチ ＝ 1 PR
- **親子Issue**: 親＝機能のまとまり（Epic。マイルストーンに紐付け・機能全体の受け入れ条件を記載）／サブ＝作業単位。**ブランチ・PRを紐づけるのはサブIssueのみ**
- 親Issueは全サブ完了後、**ユーザーが機能として動作確認してから手動クローズ**（受け入れ確認）

### ブランチ命名規則

| 種類 | 形式 | 例 |
| --- | --- | --- |
| 機能実装（サブIssue） | `phase{N}/{Issue番号}-{内容}` | `phase0/3-typespec-setup` |
| リファクタ | `refactor/{内容}` | `refactor/api-clean-architecture` |
| ドキュメント | `docs/{内容}` | `docs/architecture` |
| バグ修正 | `fix/{Issue番号}-{内容}` | `fix/21-login-redirect` |

- **phase番号は上記 Phase 0〜7 と対応**（マイルストーン M0 → `phase0` …）
- 内容は英小文字ケバブケース・短く（2〜3語まで）
- **ブランチの作成・切り替えは Claude が行う**：ユーザーが **`/start <Issue番号>`**（または「#3へ」等の指示）で Issue を選択したら、Claude が main を最新化 → 該当ブランチを作成・切り替えてから実装に入る。作業前にブランチが正しいか `git branch` で確認する
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
| 応募 | `application` | 人材→案件への応募（状態機械） |
| 本人確認 | `kyc` | eKYC。人材の本人確認書類（Phase 7） |

※ 契約（contract）・稼働報告（work_report）・検収（acceptance）・レビュー（review）は tsunagu-works 側の語彙（このリポジトリではスコープ外）

---

## 環境変数

- 実値は `.env`（gitignore）、キー名の見本は `.env.example` をコミット
- **必須**（未設定なら起動失敗）: `DATABASE_URL` / `JWT_SECRET`
- **任意**（安全なデフォルトあり）: `PORT`（既定8082） / `WEB_ORIGIN`（既定 http://localhost:3001） / `NEXT_PUBLIC_API_URL`
- 環境変数の読み取りは `api-server/` の config に集約（ハードコード禁止）
