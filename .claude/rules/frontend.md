---
description: Next.js フロントエンドを実装・修正するときのルール（Stage 1: シンプル構成）
globs:
  - "web/src/**/*.tsx"
  - "web/src/**/*.ts"
---

# フロントエンド開発ルール（Next.js App Router）

## 構成（Stage 1）

- `app/` は**ルーティングだけ**。ルートグループは**アクセス制御の境界**で3分割: `(public)`=誰でも／`(guest)`=未ログイン専用（ログイン済みは/projectsへ）／`(authenticated)`=要ログイン（未ログインは/loginへ）。page.tsx は薄く
- 部品は `components/`、ロジックは `hooks/`、**API呼び出しは必ず `lib/`**（コンポーネント・フックに fetch を直接書かない）
- `lib/` は server用（`next/headers` を使うもの）と client用（BFF/Server Action経由）を**ファイルで分離**（todo-app の型）

## Server / Client の境界（todo-app・react-basics の型）

- 既定はServer Component。`"use client"` は**必要な葉だけ**（push down）
- 読み取り＝RSCでサーバー取得、書き込み＝サーバー経由でGoへ（トークンはhttpOnly Cookie、クライアントは触らない）
- サーバー専用モジュール（`next/headers` 等）をクライアントから import しない

## UI・状態

- 状態は「画面に出す値=useState / 出さない値=useRef / サーバー状態=（後段でTanStack Query）」の使い分けを意識
- リストの key は安定した一意ID。state の配列は破壊せず新配列で置き換える
- ローディング/エラーは `loading.tsx` / `error.tsx`（境界）と、書き込み系はフックの error 状態で表示
- 破壊的操作（削除等）は確認を挟む。処理中はボタンを無効化

## 将来のリファクタ（R2/R3）に備えて

- ファイル・関数名は**ドメイン語彙**で付ける（projects / applications / contracts …）
- **命名規則**: 自作コンポーネントのファイルは PascalCase（`SignupForm.tsx`）。`components/ui/`（shadcn生成物）は CLI 準拠の kebab-case のまま／lib・hooks は camelCase（`authClient.ts` / `useAuth.ts`）
- 参考ゴール構成: next-app-router-architecture（features / external / Container-Presenter）
- 学び・ハマりは `docs/フロントエンド.md` へ
