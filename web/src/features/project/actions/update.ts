'use server'

import { redirect } from 'next/navigation'

import { updateProject } from '@/external/handler/project'

import type { ProjectFormOutput } from '../schemas/create'

// 編集の保存。成功したら一覧へ戻る（戻り値があるのは失敗時だけ）。
// id の所有チェックはサーバー側（SQL の WHERE company_id）が行う
export async function updateProjectAction(
  id: number,
  data: ProjectFormOutput,
): Promise<{ error: string }> {
  const result = await updateProject(id, data)
  if (!result.ok) {
    return { error: result.error }
  }
  redirect('/company/projects')
}
