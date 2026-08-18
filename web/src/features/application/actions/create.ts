'use server'

import { redirect } from 'next/navigation'

import { createApplication } from '@/external/handler/application'

// 応募する。成功したら応募一覧へ（戻り値があるのは失敗時だけ・ログイン/編集と同じ規約）。
// talent_id はサーバー側でトークンから解決される（入力に含めない）
export async function createApplicationAction(
  projectId: number,
  message: string,
): Promise<{ error: string }> {
  const result = await createApplication({
    project_id: projectId,
    message: message === '' ? undefined : message,
  })
  if (!result.ok) {
    return { error: result.error }
  }
  redirect('/talent/applications')
}
