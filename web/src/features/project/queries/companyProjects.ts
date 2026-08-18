import { queryOptions } from '@tanstack/react-query'

import { fetchMyProject, fetchMyProjects } from '../actions/read'

// クエリキーの階層: ['project', 'company', ...]。
// 書き込み後は invalidateQueries({ queryKey: companyProjectKeys.all }) で
// 一覧・詳細をまとめて無効化できる（前方一致）
export const companyProjectKeys = {
  all: ['project', 'company'] as const,
  list: ['project', 'company', 'list'] as const,
  detail: (id: number) => ['project', 'company', 'detail', id] as const,
}

// handler の Result をここで解いて「データ or throw」に変換する。
// useQuery はこの throw を error 状態として拾う（メッセージは handler が整えたもの）
export const companyProjectsQuery = queryOptions({
  queryKey: companyProjectKeys.list,
  queryFn: async () => {
    const result = await fetchMyProjects()
    if (!result.ok) {
      throw new Error(result.error)
    }
    return result.data
  },
})

export const companyProjectQuery = (id: number) =>
  queryOptions({
    queryKey: companyProjectKeys.detail(id),
    queryFn: async () => {
      const result = await fetchMyProject(id)
      if (!result.ok) {
        throw new Error(result.error)
      }
      return result.data
    },
  })
