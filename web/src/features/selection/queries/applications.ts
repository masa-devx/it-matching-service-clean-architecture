import { queryOptions } from '@tanstack/react-query'

import { fetchApplications } from '../actions/read'

// キー階層: ['selection', 'list', projectId]。案件ごとにキャッシュが分かれ、
// 操作後は selectionKeys.all の前方一致でまとめて無効化できる
export const selectionKeys = {
  all: ['selection'] as const,
  list: (projectId: number) => ['selection', 'list', projectId] as const,
}

export const applicationsQuery = (projectId: number) =>
  queryOptions({
    queryKey: selectionKeys.list(projectId),
    queryFn: async () => {
      const result = await fetchApplications(projectId)
      if (!result.ok) {
        throw new Error(result.error)
      }
      return result.data
    },
  })
