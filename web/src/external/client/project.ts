import 'server-only'

import { projectsCreate } from '@repo/api-client/company/generated/endpoints'
import type { TsunaguWorksProjectCreateInput } from '@repo/api-client/company/generated/models'

import { authHeaders } from './auth'

export const projectClient = {
  create: async (input: TsunaguWorksProjectCreateInput) =>
    projectsCreate(input, { headers: await authHeaders() }),
}
