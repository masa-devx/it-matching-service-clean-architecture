import { dehydrate, HydrationBoundary } from '@tanstack/react-query'
import Link from 'next/link'
import { redirect } from 'next/navigation'

import { Button } from '@/components/ui/button'

import { CompanyMeCard } from '@/features/auth/components/client/CompanyMeCard'
import { companyMeQuery } from '@/features/auth/queries/companyMe'
import { getQueryClient } from '@/lib/query'

// RSC で prefetch → dehydrate → HydrationBoundary の型:
// サーバーで温めたキャッシュをクライアントの QueryClient が引き継ぐため、
// クライアント側は初回からデータありで描画される（ローディングのちらつきが無い）
export default async function Page() {
  const queryClient = getQueryClient()

  // fetchQuery は取得とキャッシュ登録を同時に行う（ガードで値も使うため prefetchQuery でなくこちら）
  const me = await queryClient.fetchQuery(companyMeQuery)
  if (!me) {
    redirect('/company/login')
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">企業ダッシュボード</h1>

      <HydrationBoundary state={dehydrate(queryClient)}>
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
