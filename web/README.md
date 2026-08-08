# web — Next.js フロントエンド設計（出発点）

Next.js（App Router）+ TypeScript + Tailwind CSS v4 + shadcn/ui。tsunagu-works MVP の **lib集約構成**を、リファクタの出発点としてそのまま保持している。

> ⚠️ このディレクトリは **Before（リファクタ対象）**。[後継リポジトリ設計プラン](../docs/後継リポジトリ設計プラン.md)に沿って、**company / talent の2アプリ**（`apps/{company,talent}` ＋ `packages/`）へ段階的に作り替えていく。目標構成は末尾の[ロードマップ](#進化のロードマップ設計プランへ移行)へ。

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
[Go API :8081]
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

| 選定                                           | 理由（詳細は docs/フロントエンド.md の判断ログ）                                                                                                |
| ---------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| **REST + 素のfetch（axios不使用）**            | Next.jsのキャッシュ・revalidateはfetch拡張前提。axiosの利点（インターセプタ）はlib/の薄いラッパーで代替                                         |
| **GraphQL不使用**                              | 「複数クライアント×複雑データ形状×専任チーム」が揃って初めて回収できる投資。型安全は後継設計の **TypeSpec → OpenAPI → orval 生成**で獲得する    |
| **BFFはRoute Handler（Server Actions不使用）** | 仕組みを手で書いて理解する学習優先。読み取りはRSC直fetchでBFF不要のため増殖しない。後継設計では **Server Actions を採用**し、両方式を比較して語る |
| **React Hook Form 後回し**                     | 認証フォームは2〜3欄で検証の正はサーバー側。useState+HTML標準検証で十分。後継設計では **RHF ＋ 生成Zod**（orval）の組み合わせを Phase 0 で検証する |
| **shadcn/ui**                                  | npm依存でなくコードのコピーイン＝自分のコードとして改造できる。トークン（--primary等）の上書きで全体テーマ変更                                  |
| **Tailwind v4**                                | 設定はCSS内 `@theme`。ユーティリティファースト                                                                                                  |
| **Vitest + Testing Library**                   | 「ユーザーから見える振る舞い」でテスト（getByRole/getByLabelText）。a11yとテスト容易性が同じ方向を向く                                          |

## テスト

- `npm run test`（CI と同一）／`npm run test:watch`（開発時）
- lib/ に通信を集約してあるため、**lib関数1つの vi.mock でコンポーネントをテスト**できる（`LoginForm.test.tsx` が見本）

## 進化のロードマップ（設計プランへ移行）

旧計画（R2: features分割 → R3: external層）は、[後継リポジトリ設計プラン](../docs/後継リポジトリ設計プラン.md)に**置き換えられた**。1アプリ内のロール分岐をやめ、**company / talent の2アプリ（Turborepo）**に分割する。

```
現在: web/ 1アプリ（app / components / hooks / lib）
  ↓
目標: apps/{company|talent}/src/
      ├─ features/{domain}/   # components(server/client) / hooks(TanStack Query) /
      │                       #   queries / actions(Server Actions) / schemas(生成Zodの加工)
      ├─ external/            # データ取得の境界: handler（入口）→ client（Go API・server-only）
      └─ middleware.ts
      ※ 型・Fetch Client・Zod は packages/api-client（orval 生成）から供給される
```

現行コードには目標構成への「切り込み線」が入っている：

| 現行コード | 移行先 |
| --- | --- |
| `lib/api.ts`（Go API 呼び出し・server-only） | `external/client/` |
| `lib/authClient.ts`（クライアント→BFF） | `features/{domain}/actions/`（Server Actions 化） |
| `app/api/auth/*`（Route Handler の BFF） | Server Actions ＋ `external/handler/` |
| `hooks/`（mutation・ローディング/エラー集約） | `features/{domain}/hooks/`（TanStack Query） |
| `lib/types.ts`（手書きの共有型） | `packages/api-client`（orval 生成） |
| ルートグループによるロール分岐 | **アプリ境界そのもの**（company / talent 分割） |

- 参考ゴール構成: [next-app-router-architecture](https://github.com/YukiOnishi1129/next-app-router-architecture)
