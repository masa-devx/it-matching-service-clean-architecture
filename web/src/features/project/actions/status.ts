'use server'

import {
  changeProjectStatus,
  type ProjectStatusAction,
} from '@/external/handler/project'

// 状態変更（publish / unpublish / close）。redirect せず Result を返す:
// 遷移後の画面更新は呼び出し側の invalidateQueries が担う（キャッシュの一貫性）
export async function changeProjectStatusAction(
  id: number,
  action: ProjectStatusAction,
) {
  return changeProjectStatus(id, action)
}
