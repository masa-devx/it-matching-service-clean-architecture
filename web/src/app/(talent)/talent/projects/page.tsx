import { dehydrate, HydrationBoundary } from '@tanstack/react-query'

import { ProjectSearchList } from '@/features/projectSearch/components/client/ProjectSearchList'
import { SearchForm } from '@/features/projectSearch/components/client/SearchForm'
import { parseSearchFilters } from '@/features/projectSearch/filters'
import { projectSearchQuery } from '@/features/projectSearch/queries/projects'
import { getQueryClient } from '@/lib/query'

// searchParams も params と同様 Promise（Next 15+）。
// URL → filters → prefetchInfiniteQuery（1ページ目だけサーバーで温める）→ Hydration
export default async function Page({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>
}) {
  const filters = parseSearchFilters(await searchParams)

  const queryClient = getQueryClient()
  await queryClient.prefetchInfiniteQuery(projectSearchQuery(filters))

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">案件を探す</h1>

      {/* key で条件変更時にフォームを作り直す（uncontrolled の defaultValue は再マウント時にしか反映されないため） */}
      <SearchForm key={JSON.stringify(filters)} defaultValues={filters} />

      <HydrationBoundary state={dehydrate(queryClient)}>
        <ProjectSearchList filters={filters} />
      </HydrationBoundary>
    </div>
  )
}
