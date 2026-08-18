import 'server-only'

import {
  applicationsAccept,
  applicationsCreate,
  applicationsDecline,
  applicationsList,
  applicationsWithdraw,
} from '@repo/api-client/talent/generated/endpoints'
import type { TsunaguWorksApplicationCreateInput } from '@repo/api-client/talent/generated/models'

import { authHeaders } from './auth'

// 応募（talent 視点: 出す・取り下げる・応諾する）のクライアント
export const applicationClient = {
  create: async (input: TsunaguWorksApplicationCreateInput) =>
    applicationsCreate(input, { headers: await authHeaders() }),
  list: async () => applicationsList({ headers: await authHeaders() }),
  withdraw: async (id: number) =>
    applicationsWithdraw(id, { headers: await authHeaders() }),
  accept: async (id: number) =>
    applicationsAccept(id, { headers: await authHeaders() }),
  decline: async (id: number) =>
    applicationsDecline(id, { headers: await authHeaders() }),
}
