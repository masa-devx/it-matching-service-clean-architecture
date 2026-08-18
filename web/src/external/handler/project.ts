import 'server-only'

import type {
  TsunaguWorksProject,
  TsunaguWorksProjectCreateInput,
  TsunaguWorksProjectUpdateInput,
} from '@repo/api-client/company/generated/models'

import { projectClient } from '../client/project'

// handler は通信結果を「画面が分岐できる形（Result）」に整える。
// API のエラーメッセージ（404/409 の文言は Go 側が一次情報）はそのまま通し、
// 通信不能や欠落時だけフォールバック文言を足す

type Result<T> = { ok: true; data: T } | { ok: false; error: string }

const CONNECT_ERROR = 'サーバーに接続できませんでした'

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
    return { ok: false, error: CONNECT_ERROR }
  }
}

export async function listMyProjects(): Promise<Result<TsunaguWorksProject[]>> {
  try {
    const res = await projectClient.list()
    if (res.status !== 200) {
      return { ok: false, error: res.data.error ?? '一覧の取得に失敗しました' }
    }
    return { ok: true, data: res.data.projects }
  } catch {
    return { ok: false, error: CONNECT_ERROR }
  }
}

export async function getMyProject(
  id: number,
): Promise<Result<TsunaguWorksProject>> {
  try {
    const res = await projectClient.get(id)
    if (res.status !== 200) {
      return { ok: false, error: res.data.error ?? '案件が見つかりません' }
    }
    return { ok: true, data: res.data }
  } catch {
    return { ok: false, error: CONNECT_ERROR }
  }
}

export async function updateProject(
  id: number,
  input: TsunaguWorksProjectUpdateInput,
): Promise<Result<TsunaguWorksProject>> {
  try {
    const res = await projectClient.update(id, input)
    if (res.status !== 200) {
      return { ok: false, error: res.data.error ?? '案件の更新に失敗しました' }
    }
    return { ok: true, data: res.data }
  } catch {
    return { ok: false, error: CONNECT_ERROR }
  }
}

// 状態変更は3操作とも「id を受けて更新後の案件を返す」同形なので1つの入口に束ねる。
// 409（遷移不可）のメッセージは Go 側が「可能な遷移先つき」で返すためそのまま画面へ
export type ProjectStatusAction = 'publish' | 'unpublish' | 'close'

export async function changeProjectStatus(
  id: number,
  action: ProjectStatusAction,
): Promise<Result<TsunaguWorksProject>> {
  try {
    const res = await projectClient[action](id)
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
