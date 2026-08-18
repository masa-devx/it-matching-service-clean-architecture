import { dehydrate, HydrationBoundary } from '@tanstack/react-query'
import Link from 'next/link'
import { redirect } from 'next/navigation'

import { CompanyMeCard } from '@/features/auth/components/client/CompanyMeCard'
import { LogoutButton } from '@/features/auth/components/client/LogoutButton'
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
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">企業ダッシュボード</h1>
        <LogoutButton role="company" />
      </div>

      <HydrationBoundary state={dehydrate(queryClient)}>
        <CompanyMeCard />
      </HydrationBoundary>

      <Link
        href="/company/projects/new"
        className="inline-block rounded bg-blue-600 px-4 py-2 text-white"
      >
        案件を作成する
      </Link>
    </div>
  )
}
