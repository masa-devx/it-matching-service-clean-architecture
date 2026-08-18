import { Skeleton } from '@/components/ui/skeleton'

// ページ遷移中のスケルトン（RSC の prefetch が終わるまでの間に表示される境界）。
// 一覧1件分の形（タイトル行+メタ2行）を模して、完了時のレイアウト差を小さくする
export default function Loading() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-8 w-40" />
      <div className="space-y-3">
        {Array.from({ length: 5 }, (_, i) => (
          <div key={i} className="space-y-2 rounded-lg border p-4">
            <Skeleton className="h-5 w-2/3" />
            <Skeleton className="h-4 w-1/2" />
            <Skeleton className="h-4 w-1/3" />
          </div>
        ))}
      </div>
    </div>
  )
}
