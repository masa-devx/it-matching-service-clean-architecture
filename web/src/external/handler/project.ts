import 'server-only'

import type {
  TsunaguWorksProject,
  TsunaguWorksProjectCreateInput,
} from '@repo/api-client/company/generated/models'

import { projectClient } from '../client/project'

type CreateResult =
  { ok: true; project: TsunaguWorksProject } | { ok: false; error: string }

export async function createProject(
  input: TsunaguWorksProjectCreateInput,
): Promise<CreateResult> {
  try {
    const res = await projectClient.create(input)
    if (res.status !== 201) {
      return { ok: false, error: res.data.error ?? '案件の作成に失敗しました' }
    }
    return { ok: true, project: res.data }
  } catch {
    return { ok: false, error: 'サーバーに接続できませんでした' }
  }
}
