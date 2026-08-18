import 'server-only'

import type {
  TsunaguWorksApplication,
  TsunaguWorksApplicationCreateInput,
} from '@repo/api-client/talent/generated/models'

import { applicationClient } from '../client/application'

type Result<T> = { ok: true; data: T } | { ok: false; error: string }

const CONNECT_ERROR = 'サーバーに接続できませんでした'

// 404（未公開案件）・409（二重応募）の文言は Go を一次情報として素通し
export async function createApplication(
  input: TsunaguWorksApplicationCreateInput,
): Promise<Result<TsunaguWorksApplication>> {
  try {
    const res = await applicationClient.create(input)
    if (res.status !== 201) {
      return { ok: false, error: res.data.error ?? '応募に失敗しました' }
    }
    return { ok: true, data: res.data }
  } catch {
    return { ok: false, error: CONNECT_ERROR }
  }
}

export async function listMyApplications(): Promise<
  Result<TsunaguWorksApplication[]>
> {
  try {
    const res = await applicationClient.list()
    if (res.status !== 200) {
      return { ok: false, error: res.data.error ?? '一覧の取得に失敗しました' }
    }
    return { ok: true, data: res.data.applications }
  } catch {
    return { ok: false, error: CONNECT_ERROR }
  }
}

// withdraw / accept / decline は同形なので1つの入口に束ねる。
// 409（遷移競合）の「現在: x」つき文言は Go の遷移表由来なのでそのまま画面へ
export type ApplicationAction = 'withdraw' | 'accept' | 'decline'

export async function changeApplication(
  applicationId: number,
  action: ApplicationAction,
): Promise<Result<TsunaguWorksApplication>> {
  try {
    const res = await applicationClient[action](applicationId)
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
