import { Skeleton } from '@/components/ui/skeleton'
import { ProjectListSkeleton } from '@/components/ProjectListSkeleton'

// 一覧セグメント専用のローディング。検索・ページ送りのたびに表示される
export default function Loading() {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-4 w-32" />
      </div>
      <Skeleton className="h-32 w-full rounded-lg" />
      <ProjectListSkeleton />
    </div>
  )
}
