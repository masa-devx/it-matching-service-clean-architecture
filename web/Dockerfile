# web（Next.js）の本番用イメージ（ADR-0011: Cloud Run 統一）。
# ビルドコンテキストはリポジトリルート:
#   docker build -f web/Dockerfile -t tsunagu-web .
# ignore は web/Dockerfile.dockerignore（Dockerfile ごとの ignore・BuildKit）が効く

FROM node:24-slim AS builder
WORKDIR /src

# packageManager フィールド（ルート package.json）の版で pnpm を有効化
RUN corepack enable

# 依存解決に必要なマニフェストだけ先にコピー（レイヤーキャッシュ: lockfile が変わらない限り install をスキップ）。
# api-server/package.json も必要: pnpm workspace の一員なので frozen-lockfile の照合対象になる
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml turbo.json ./
COPY web/package.json web/
COPY api-server/package.json api-server/
COPY packages/spec/package.json packages/spec/
COPY packages/api-client/package.json packages/api-client/
COPY packages/typescript-config/package.json packages/typescript-config/
RUN pnpm install --frozen-lockfile

# ソースを載せて生成チェーン込みでビルド（spec → api-client → web が turbo の依存グラフで連鎖）
COPY web/ web/
COPY packages/ packages/
RUN pnpm turbo build --filter=web

# ランタイム: standalone 出力（実行に必要な最小 node_modules + server.js の自己完結セット）だけを載せる。
# distroless/nodejs は Node 版の追従が遅いため node:24-slim を使う（.nvmrc と同一メジャー）
FROM node:24-slim
WORKDIR /app
ENV NODE_ENV=production
# Cloud Run はコンテナ外から 0.0.0.0 宛てに接続する（既定の localhost バインドでは届かない）
ENV HOSTNAME=0.0.0.0

# モノレポの standalone はリポジトリ構造（web/server.js + ルート node_modules）を保って出力される
COPY --from=builder /src/web/.next/standalone ./
# 静的アセットは standalone に含まれない（CDN 配信を想定した設計のため）→ 明示的に併置する
COPY --from=builder /src/web/.next/static ./web/.next/static

EXPOSE 3000
# PORT は Cloud Run が注入し、Next standalone サーバーはそれを尊重する
CMD ["node", "web/server.js"]
