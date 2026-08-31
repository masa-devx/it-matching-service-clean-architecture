# web — Next.js フロントエンド設計（出発点）

Next.js（App Router）+ TypeScript + Tailwind CSS v4 + shadcn/ui。tsunagu-works MVP の **lib集約構成**を、リファクタの出発点としてそのまま保持している。

> ℹ️ このディレクトリは tsunagu-works MVP を出発点に、**features / external 構成へ作り替え済み**（[ADR-0006](../docs/adr/0006-single-frontend-after-only.md) で 1 アプリ維持を決定）。移行の経緯は末尾の[ロードマップ](#進化のロードマップ移行完了の記録)へ。本文の「lib 集約」の記述は出発点（Stage 1）の設計記録として残している。

## 構成とディレクトリ責務

```
web/src/
├── app/                       # ルーティングだけ（page は薄く）
│   ├── (public)/              # 誰でも閲覧可（LP）
│   ├── (guest)/               # 未ログイン専用（login / signup。ログイン済みは /projects へ）
│   ├── (authenticated)/       # 要ログイン（未ログインは /login へ）＋共通ヘッダー
│   │   ├── error.tsx          # エラー境界（再試行つき・ヘッダーは生き残る）
│   │   └── loading.tsx        # Suspenseフォールバック（スケルトン）
│   ├── api/auth/*             # BFF（Route Handler）: Go呼び出し＋httpOnly Cookie変換
│   └── not-found.tsx          # 404
├── components/                # 自作UI部品（PascalCase）
│   └── ui/                    # shadcn生成物（kebab-case・CLI管理）
├── hooks/                     # カスタムフック（camelCase）
├── lib/                       # ★fetchの唯一の置き場（コンポーネントに書かない）
│   ├── api.ts                 # Go API呼び出し（server-only・判別可能ユニオンで結果を返す）
│   ├── auth.ts                # getCurrentUser: Cookie → GET /me 検証（server-only）
│   ├── authCookie.ts          # httpOnly Cookie の set/get/delete（server-only）
│   └── authClient.ts          # クライアント→BFF呼び出し（トークンを知らない）
└── test/setup.ts              # Vitest セットアップ
```

## 認証・通信の構成（BFF）

```
[ブラウザ]
   │ 同一オリジンのみ（トークンに触れない）
   ▼
[Next.jsサーバー]
   ├─ RSC: cookies() → Bearer付きで Go を直接fetch（読み取り・中継ルート不要）
   └─ Route Handler(/api/auth/*): Go呼び出し → token を httpOnly Cookie に変換（書き込み）
   ▼
[Go API :8082]
```

- **トークンはブラウザに一切渡さない**（httpOnly Cookie・XSS対策）
- ガードは各ルートグループの layout.tsx に集約（(authenticated)=未ログイン弾き / (guest)=ログイン済み弾き）
- 認証状態の判定は Cookie 有無でなく **Go の GET /me を単一の真実**として使う

## 設計の型（このリポジトリの規約）

- **既定は Server Component**。`"use client"` は必要な葉だけ（push down）
- **読み取り=RSCでサーバーfetch／書き込み=BFF経由**。fetch は lib/ のみ
- lib/ は **server用（server-only付き）と client用をファイルで分離**
- 命名: コンポーネント=PascalCase／shadcn生成物=kebab-case／lib・hooks=camelCase
- フォーム: ラベル必須・エラーはフィールド直下＋`role="alert"`・送信中はボタン無効
- ダークモード非対応（`dark:` を書かない）。デザイントークンは globals.css の `:root`（docs/デザインシステム.md が正）

## 技術選定と理由

| 選定                                           | 理由（詳細は docs/フロントエンド.md の判断ログ）                                                                                                   |
| ---------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| **REST + 素のfetch（axios不使用）**            | Next.jsのキャッシュ・revalidateはfetch拡張前提。axiosの利点（インターセプタ）はlib/の薄いラッパーで代替                                            |
| **GraphQL不使用**                              | 「複数クライアント×複雑データ形状×専任チーム」が揃って初めて回収できる投資。型安全は後継設計の **TypeSpec → OpenAPI → orval 生成**で獲得する       |
| **BFFはRoute Handler（Server Actions不使用）** | 仕組みを手で書いて理解する学習優先。読み取りはRSC直fetchでBFF不要のため増殖しない。後継設計では **Server Actions を採用**し、両方式を比較して語る  |
| **React Hook Form 後回し**                     | 認証フォームは2〜3欄で検証の正はサーバー側。useState+HTML標準検証で十分。後継設計では **RHF ＋ 生成Zod**（orval）の組み合わせを Phase 0 で検証する |
| **shadcn/ui**                                  | npm依存でなくコードのコピーイン＝自分のコードとして改造できる。トークン（--primary等）の上書きで全体テーマ変更                                     |
| **Tailwind v4**                                | 設定はCSS内 `@theme`。ユーティリティファースト                                                                                                     |
| **Vitest + Testing Library**                   | 「ユーザーから見える振る舞い」でテスト（getByRole/getByLabelText）。a11yとテスト容易性が同じ方向を向く                                             |

## テスト

- `npm run test`（CI と同一）／`npm run test:watch`（開発時）
- lib/ に通信を集約してあるため、**lib関数1つの vi.mock でコンポーネントをテスト**できる（`LoginForm.test.tsx` が見本）

## 進化のロードマップ（移行完了の記録）

旧計画（R2: features分割 → R3: external層）→ 設計プランの **2アプリ分割**案 → [ADR-0006](../docs/adr/0006-single-frontend-after-only.md) で **1アプリ維持**に確定、という経緯を辿り、Phase 2〜4（#32・#44〜#46・#55〜#60）で移行を**完了**した。

```
出発点: web/ 1アプリ（app / components / hooks / lib）
   ↓ ADR-0006（2アプリ分割は取り下げ・ルートグループでロール分岐）
現在:   web/src/
        ├─ app/(public|guest|company|talent)/  # ルーティングと認可の境界だけ
        ├─ features/{auth,project,projectSearch,application,selection}/
        │    # components(server/client) / actions(Server Actions) / queries(TanStack Query) / schemas
        ├─ external/            # handler（入口）→ client（Go API・server-only）
        └─ components/          # ドメインを知らない共有部品
        ※ 型・Fetch Client・Zod は packages/api-client（orval 生成）から供給
```

移行の対応表（すべて移行済み）:

| 出発点のコード                                | 移行先（現在）                                      |
| --------------------------------------------- | --------------------------------------------------- |
| `lib/api.ts`（Go API 呼び出し・server-only）  | `external/client/`                                  |
| `lib/authClient.ts`（クライアント→BFF）       | `features/{domain}/actions/`（Server Actions）      |
| `app/api/auth/*`（Route Handler の BFF）      | Server Actions ＋ `external/handler/`               |
| `hooks/`（mutation・ローディング/エラー集約） | `features/{domain}/queries/`（TanStack Query）      |
| `lib/types.ts`（手書きの共有型）              | `packages/api-client`（orval 生成）                 |
| ルートグループによるロール分岐                | **維持**（(guest)/(company)/(talent) ＝認可の境界） |

- 参考ゴール構成: [next-app-router-architecture](https://github.com/YukiOnishi1129/next-app-router-architecture)
