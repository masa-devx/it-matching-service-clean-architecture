import { Skeleton } from '@/components/ui/skeleton'

// 実画面に近い形のスケルトンにすることで、読み込み完了時のレイアウトのズレを防ぎ、
// 体感速度を上げる（汎用の四角を並べるだけだと「何が来るか」が伝わらない）
export function ProjectListSkeleton({ count = 5 }: { count?: number }) {
  return (
    <ul className="flex flex-col gap-3" aria-hidden="true">
      {Array.from({ length: count }).map((_, i) => (
        <li
          key={i}
          className="flex flex-col gap-3 rounded-lg border bg-card p-4"
        >
          <div className="flex flex-col gap-1">
            <Skeleton className="h-5 w-2/3" />
            <Skeleton className="h-4 w-32" />
          </div>
          <Skeleton className="h-6 w-40" />
          <div className="flex gap-2">
            <Skeleton className="h-5 w-16 rounded-full" />
            <Skeleton className="h-5 w-20 rounded-full" />
          </div>
        </li>
      ))}
    </ul>
  )
}
