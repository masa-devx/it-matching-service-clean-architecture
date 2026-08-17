import 'server-only'

import {
  authLogin as companyAuthLogin,
  authMe as companyAuthMe,
  authSignup as companyAuthSignup,
} from '@repo/api-client/company/generated/endpoints'
import type {
  TsunaguWorksCompanySignupInput,
  TsunaguWorksLoginInput,
} from '@repo/api-client/company/generated/models'
import {
  authLogin as talentAuthLogin,
  authMe as talentAuthMe,
  authSignup as talentAuthSignup,
} from '@repo/api-client/talent/generated/endpoints'
import type { TsunaguWorksTalentSignupInput } from '@repo/api-client/talent/generated/models'

import { getSessionToken } from './session'

// 要認証エンドポイント用に Bearer を付ける。
// トークンが無ければ付けずに送り、判定は Go 側の 401 に任せる（真実は常にサーバー）
export const authHeaders = async (): Promise<HeadersInit> => {
  const token = await getSessionToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export const companyClient = {
  signup: (input: TsunaguWorksCompanySignupInput) => companyAuthSignup(input),
  login: (input: TsunaguWorksLoginInput) => companyAuthLogin(input),
  me: async () => companyAuthMe({ headers: await authHeaders() }),
}

export const talentClient = {
  signup: (input: TsunaguWorksTalentSignupInput) => talentAuthSignup(input),
  login: (input: TsunaguWorksLoginInput) => talentAuthLogin(input),
  me: async () => talentAuthMe({ headers: await authHeaders() }),
}
