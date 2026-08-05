import { Skeleton } from '@/components/ui/skeleton'

// (authenticated) 配下のページ読み込み中に自動表示される（Suspense のフォールバック）。
// 一覧画面を想定した汎用スケルトン
export default function Loading() {
  return (
    <div className="flex flex-col gap-4">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-24 w-full" />
      <Skeleton className="h-24 w-full" />
      <Skeleton className="h-24 w-full" />
    </div>
  )
}
