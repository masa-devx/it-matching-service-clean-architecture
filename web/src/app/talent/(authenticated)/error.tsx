'use client'

import { useEffect } from 'react'
import { Button } from '@/components/ui/button'

// (authenticated) 配下のページで起きたレンダリングエラーを受け止める境界。
// エラー境界はクライアント側の仕組みのため 'use client' が必須
export default function ErrorBoundary({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    // 本番想定では監視サービスへ送る場所。今は開発者向けにconsoleへ
    console.error(error)
  }, [error])

  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-4 py-16">
      <h2 className="text-xl font-bold">問題が発生しました</h2>
      <p className="text-sm text-muted-foreground">
        時間をおいて再度お試しください。解決しない場合はお問い合わせください。
      </p>
      <Button onClick={reset} className="h-11">
        再試行
      </Button>
    </div>
  )
}
