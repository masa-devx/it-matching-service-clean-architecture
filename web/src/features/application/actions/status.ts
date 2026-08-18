'use server'

import {
  changeApplication,
  type ApplicationAction,
} from '@/external/handler/application'

// 応募の遷移操作（withdraw / accept / decline）。redirect せず Result を返し、
// 画面更新は呼び出し側の invalidateQueries が担う
export async function changeApplicationAction(
  applicationId: number,
  action: ApplicationAction,
) {
  return changeApplication(applicationId, action)
}
