import 'server-only'

import {
  projectsGet,
  projectsList,
} from '@repo/api-client/talent/generated/endpoints'
import type { ProjectsListParams } from '@repo/api-client/talent/generated/models'

import { authHeaders } from './auth'

// talent 視点の案件閲覧クライアント。company 側（client/project.ts）とは別物:
// 見えるのは公開案件のみで、検索条件と seek カーソルを持つ
export const projectSearchClient = {
  list: async (params: ProjectsListParams) =>
    projectsList(params, { headers: await authHeaders() }),
  get: async (id: number) => projectsGet(id, { headers: await authHeaders() }),
}
