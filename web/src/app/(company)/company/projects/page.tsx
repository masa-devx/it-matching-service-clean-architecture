import { dehydrate, HydrationBoundary } from '@tanstack/react-query'
import Link from 'next/link'

import { Button } from '@/components/ui/button'
import { ProjectList } from '@/features/project/components/client/ProjectList'
import { companyProjectsQuery } from '@/features/project/queries/companyProjects'
import { getQueryClient } from '@/lib/query'

// 一覧はページ側で値を使わないため prefetchQuery（fetchQuery との使い分け）。
// 認証ガードは (company)/layout.tsx が担う
export default async function Page() {
  const queryClient = getQueryClient()
  await queryClient.prefetchQuery(companyProjectsQuery)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">案件管理</h1>
        <Button asChild>
          <Link href="/company/projects/new">新規作成</Link>
        </Button>
      </div>

      <HydrationBoundary state={dehydrate(queryClient)}>
        <ProjectList />
      </HydrationBoundary>
    </div>
  )
}
