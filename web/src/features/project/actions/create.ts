'use server'

import { createProject } from '@/external/handler/project'
import type { TsunaguWorksProject } from '@repo/api-client/company/generated/models'

import type { ProjectFormOutput } from '../schemas/create'

type ActionResult =
  { ok: true; project: TsunaguWorksProject } | { ok: false; error: string }

export async function createProjectAction(
  data: ProjectFormOutput,
): Promise<ActionResult> {
  return createProject(data)
}
