import { queryOptions } from '@tanstack/react-query'

import { fetchMyApplications } from '../actions/read'

// 自分の応募一覧（talent には1本しかないのでキーに引数なし）
export const applicationKeys = {
  all: ['application'] as const,
  list: ['application', 'list'] as const,
}

export const myApplicationsQuery = queryOptions({
  queryKey: applicationKeys.list,
  queryFn: async () => {
    const result = await fetchMyApplications()
    if (!result.ok) {
      throw new Error(result.error)
    }
    return result.data
  },
})
