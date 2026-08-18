import { environmentManager, QueryClient } from '@tanstack/react-query'

// QueryClient の寿命はサーバーとブラウザで変える:
// - サーバー: リクエストごとに新規作成（共有すると他人のリクエストにキャッシュが漏れる）
// - ブラウザ: シングルトン（再レンダリングで作り直すとキャッシュが消え、Suspense 中は無限ループの原因になる)
function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        // RSC で prefetch した直後にクライアントが再フェッチしないよう、
        // 「新鮮」とみなす時間を既定で持たせる（画面ごとの上書きは queryOptions 側で）
        staleTime: 30 * 1000,
      },
    },
  })
}

let browserQueryClient: QueryClient | undefined

export function getQueryClient(): QueryClient {
  if (environmentManager.isServer()) {
    return makeQueryClient()
  }
  browserQueryClient ??= makeQueryClient()
  return browserQueryClient
}
