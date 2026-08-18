import 'server-only'

import type { TsunaguWorksApplicationForCompany } from '@repo/api-client/company/generated/models'

import { selectionClient } from '../client/selection'

type Result<T> = { ok: true; data: T } | { ok: false; error: string }

const CONNECT_ERROR = 'サーバーに接続できませんでした'

// 他社の案件は API が 404 を返す（存在の有無を漏らさない）。文言は Go を一次情報として素通し
export async function listApplications(
  projectId: number,
): Promise<Result<TsunaguWorksApplicationForCompany[]>> {
  try {
    const res = await selectionClient.listApplications(projectId)
    if (res.status !== 200) {
      return { ok: false, error: res.data.error ?? '一覧の取得に失敗しました' }
    }
    return { ok: true, data: res.data.applications }
  } catch {
    return { ok: false, error: CONNECT_ERROR }
  }
}

// offer / reject は同形なので1つの入口に束ねる（#45 changeProjectStatus と同じ型）。
// 409（遷移競合）の「現在: x」つき文言は Go の遷移表由来なのでそのまま画面へ
export type SelectionAction = 'offer' | 'reject'

export async function changeSelection(
  applicationId: number,
  action: SelectionAction,
): Promise<Result<TsunaguWorksApplicationForCompany>> {
  try {
    const res = await selectionClient[action](applicationId)
    if (res.status !== 200) {
      return {
        ok: false,
        error: res.data.error ?? '現在の状態ではこの操作はできません',
      }
    }
    return { ok: true, data: res.data }
  } catch {
    return { ok: false, error: CONNECT_ERROR }
  }
}
