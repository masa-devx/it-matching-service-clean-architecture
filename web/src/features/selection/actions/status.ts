'use server'

import {
  changeSelection,
  type SelectionAction,
} from '@/external/handler/selection'

// 選考操作（offer / reject）。redirect せず Result を返し、
// 画面更新は呼び出し側の invalidateQueries が担う（#45 の型）
export async function changeSelectionAction(
  applicationId: number,
  action: SelectionAction,
) {
  return changeSelection(applicationId, action)
}
