import { Skeleton } from '@/components/ui/skeleton'

// 絞り込み・ページ送りのたびに表示される
export default function Loading() {
  return (
    <div className="flex flex-col gap-6">
      <Skeleton className="h-5 w-32" />
      <div className="flex flex-col gap-1">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-4 w-56" />
      </div>
      <Skeleton className="h-10 w-full max-w-lg rounded-full" />
      <div className="flex flex-col gap-3">
        {[0, 1, 2].map((i) => (
          <Skeleton key={i} className="h-52 w-full rounded-lg" />
        ))}
      </div>
    </div>
  )
}
