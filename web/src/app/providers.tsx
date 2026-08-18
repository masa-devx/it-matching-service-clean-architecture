'use client'

import { QueryClientProvider } from '@tanstack/react-query'
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'

import { getQueryClient } from '@/lib/query'

// app は「合成の層」: アプリ全体に効かせる Provider をここで束ねて layout に渡す。
// useState で持たない理由は lib/query.ts の getQueryClient 参照
export function Providers({ children }: { children: React.ReactNode }) {
  const queryClient = getQueryClient()

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      {/* 本番ビルドでは自動で除外される */}
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  )
}
