import { dehydrate, HydrationBoundary } from '@tanstack/react-query'
import Link from 'next/link'
import { redirect } from 'next/navigation'

import { Button } from '@/components/ui/button'
import { CompanyMeCard } from '@/features/auth/components/client/CompanyMeCard'
import { companyMeQuery } from '@/features/auth/queries/companyMe'
import { CompanyDashboardStats } from '@/features/project/components/client/CompanyDashboardStats'
import { companyProjectsQuery } from '@/features/project/queries/companyProjects'
import { getQueryClient } from '@/lib/query'

// me はガードで値も使うため fetchQuery・案件一覧（統計の元データ）は prefetchQuery。
// 統計は一覧キャッシュからの導出なので、専用の集計 API は無い
export default async function Page() {
  const queryClient = getQueryClient()

  const me = await queryClient.fetchQuery(companyMeQuery)
  if (!me) {
    redirect('/company/login')
  }
  await queryClient.prefetchQuery(companyProjectsQuery)

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">企業ダッシュボード</h1>

      <HydrationBoundary state={dehydrate(queryClient)}>
        <CompanyDashboardStats />
        <CompanyMeCard />
      </HydrationBoundary>

      <div className="flex gap-2">
        <Button asChild>
          <Link href="/company/projects">案件管理へ</Link>
        </Button>
        <Button asChild variant="outline">
          <Link href="/company/projects/new">案件を作成する</Link>
        </Button>
      </div>
    </div>
  )
}
