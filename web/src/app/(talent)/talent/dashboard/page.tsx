import { dehydrate, HydrationBoundary } from '@tanstack/react-query'
import Link from 'next/link'
import { redirect } from 'next/navigation'

import { Button } from '@/components/ui/button'
import { meTalent } from '@/external/handler/auth'
import { TalentDashboardStats } from '@/features/application/components/client/TalentDashboardStats'
import { myApplicationsQuery } from '@/features/application/queries/applications'
import { getQueryClient } from '@/lib/query'

// プロフィール表示はワンショット読み（handler 直）・応募一覧（統計の元データ）は prefetch。
// 「今やるべきこと」（オファーあり）をダッシュボードの最上部で可視化する
export default async function Page() {
  const me = await meTalent()
  if (!me) {
    redirect('/talent/login')
  }

  const queryClient = getQueryClient()
  await queryClient.prefetchQuery(myApplicationsQuery)

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">人材ダッシュボード</h1>

      <HydrationBoundary state={dehydrate(queryClient)}>
        <TalentDashboardStats />
      </HydrationBoundary>

      <dl className="space-y-2 rounded-lg border bg-card p-4">
        <div>
          <dt className="text-sm text-muted-foreground">表示名</dt>
          <dd className="font-medium">{me.display_name}</dd>
        </div>
        <div>
          <dt className="text-sm text-muted-foreground">メールアドレス</dt>
          <dd className="font-medium">{me.email}</dd>
        </div>
        <div>
          <dt className="text-sm text-muted-foreground">スキル</dt>
          <dd className="font-medium">
            {me.skills.length > 0 ? me.skills.join(' / ') : '未登録'}
          </dd>
        </div>
      </dl>

      <div className="flex gap-2">
        <Button asChild>
          <Link href="/talent/projects">案件を探す</Link>
        </Button>
        <Button asChild variant="outline">
          <Link href="/talent/applications">応募一覧</Link>
        </Button>
      </div>
    </div>
  )
}
