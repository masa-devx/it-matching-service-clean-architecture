import { dehydrate, HydrationBoundary } from '@tanstack/react-query'

import { ApplicationList } from '@/features/application/components/client/ApplicationList'
import { myApplicationsQuery } from '@/features/application/queries/applications'
import { getQueryClient } from '@/lib/query'

// 自分の応募一覧。操作（承諾・辞退・取り下げ）後の invalidate を効かせるため Query に載せる
export default async function Page() {
  const queryClient = getQueryClient()
  await queryClient.prefetchQuery(myApplicationsQuery)

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">応募一覧</h1>

      <HydrationBoundary state={dehydrate(queryClient)}>
        <ApplicationList />
      </HydrationBoundary>
    </div>
  )
}
