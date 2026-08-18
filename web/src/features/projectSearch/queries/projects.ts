import { infiniteQueryOptions } from '@tanstack/react-query'

import { fetchProjects } from '../actions/read'

// 検索条件。フォーム・URL・クエリキーの3者で共有する形
export type ProjectSearchFilters = {
  skills: string[]
  remoteOk?: boolean
  minHourlyRate?: number
}

const PAGE_SIZE = 20

export const projectSearchKeys = {
  all: ['projectSearch'] as const,
  // 検索条件をキーに含める: 条件が変わればキャッシュが切り替わり、
  // 同じ条件に戻れば読み込み済みのページがそのまま出る
  list: (filters: ProjectSearchFilters) =>
    ['projectSearch', 'list', filters] as const,
}

// useInfiniteQuery 用の定義。seek の next_cursor がそのまま次ページの鍵:
// getNextPageParam が返した値が、次の queryFn 呼び出しの pageParam に入る
// （null = 次ページなし → undefined に潰すと hasNextPage が false になる）
export const projectSearchQuery = (filters: ProjectSearchFilters) =>
  infiniteQueryOptions({
    queryKey: projectSearchKeys.list(filters),
    queryFn: async ({ pageParam }) => {
      const result = await fetchProjects({
        cursor: pageParam,
        limit: PAGE_SIZE,
        skills: filters.skills.length > 0 ? filters.skills : undefined,
        remote_ok: filters.remoteOk,
        min_hourly_rate: filters.minHourlyRate,
      })
      if (!result.ok) {
        throw new Error(result.error)
      }
      return result.data
    },
    initialPageParam: undefined as number | undefined,
    getNextPageParam: (lastPage) => lastPage.next_cursor ?? undefined,
  })
