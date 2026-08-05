'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import { Button } from '@/components/ui/button'

// OFFSET ベースのページャ。ページ番号を URL に持たせるため、
// リロード・共有・ブラウザバックで同じページに戻れる
export function Pagination({
  currentPage,
  total,
  perPage,
}: {
  currentPage: number
  total: number
  perPage: number
}) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const lastPage = Math.max(1, Math.ceil(total / perPage))

  if (lastPage <= 1) {
    return null
  }

  function goTo(page: number) {
    // 既存の検索条件を保ったまま page だけ差し替える
    const params = new URLSearchParams(searchParams.toString())
    params.set('page', String(page))
    router.push(`?${params.toString()}`)
  }

  return (
    <nav
      className="flex items-center justify-center gap-4"
      aria-label="ページ送り"
    >
      <Button
        variant="outline"
        className="h-11"
        disabled={currentPage <= 1}
        onClick={() => goTo(currentPage - 1)}
      >
        前へ
      </Button>
      <span className="text-sm tabular-nums text-muted-foreground">
        {currentPage} / {lastPage}
      </span>
      <Button
        variant="outline"
        className="h-11"
        disabled={currentPage >= lastPage}
        onClick={() => goTo(currentPage + 1)}
      >
        次へ
      </Button>
    </nav>
  )
}
