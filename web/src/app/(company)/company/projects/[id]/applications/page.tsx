import { dehydrate, HydrationBoundary } from '@tanstack/react-query'
import Link from 'next/link'
import { notFound } from 'next/navigation'

import { Button } from '@/components/ui/button'
import { getMyProject } from '@/external/handler/project'
import { ApplicationList } from '@/features/selection/components/client/ApplicationList'
import { applicationsQuery } from '@/features/selection/queries/applications'
import { getQueryClient } from '@/lib/query'

// 案件の所有確認（他社・不存在は404）を先に行い、タイトルをヘッダーに出す。
// 応募一覧そのものは prefetch → Hydration でクライアントへ（操作後の invalidate を効かせるため）
export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const projectId = Number(id)
  if (!Number.isInteger(projectId) || projectId <= 0) {
    notFound()
  }

  const result = await getMyProject(projectId)
  if (!result.ok) {
    notFound()
  }

  const queryClient = getQueryClient()
  await queryClient.prefetchQuery(applicationsQuery(projectId))

  return (
    <div className="space-y-6">
      <div>
        <Button asChild variant="ghost" size="sm">
          <Link href="/company/projects">← 案件管理へ</Link>
        </Button>
      </div>

      <div className="space-y-1">
        <h1 className="text-2xl font-bold">応募一覧</h1>
        <p className="text-sm text-muted-foreground">{result.data.title}</p>
      </div>

      <HydrationBoundary state={dehydrate(queryClient)}>
        <ApplicationList projectId={projectId} />
      </HydrationBoundary>
    </div>
  )
}
