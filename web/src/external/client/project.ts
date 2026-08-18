import 'server-only'

import {
  projectsClose,
  projectsCreate,
  projectsGet,
  projectsList,
  projectsPublish,
  projectsUnpublish,
  projectsUpdate,
} from '@repo/api-client/company/generated/endpoints'
import type {
  TsunaguWorksProjectCreateInput,
  TsunaguWorksProjectUpdateInput,
} from '@repo/api-client/company/generated/models'

import { authHeaders } from './auth'

// 生成クライアントの薄い束ね: 認証ヘッダーの付与だけを一元化する。
// エラーの解釈（ステータス→メッセージ）は handler の責務
export const projectClient = {
  create: async (input: TsunaguWorksProjectCreateInput) =>
    projectsCreate(input, { headers: await authHeaders() }),
  list: async () => projectsList({ headers: await authHeaders() }),
  get: async (id: number) => projectsGet(id, { headers: await authHeaders() }),
  update: async (id: number, input: TsunaguWorksProjectUpdateInput) =>
    projectsUpdate(id, input, { headers: await authHeaders() }),
  publish: async (id: number) =>
    projectsPublish(id, { headers: await authHeaders() }),
  unpublish: async (id: number) =>
    projectsUnpublish(id, { headers: await authHeaders() }),
  close: async (id: number) =>
    projectsClose(id, { headers: await authHeaders() }),
}
