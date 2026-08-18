import 'server-only'

import type {
  ProjectsListParams,
  TsunaguWorksProject,
  TsunaguWorksProjectPage,
} from '@repo/api-client/talent/generated/models'

import { projectSearchClient } from '../client/projectSearch'

type Result<T> = { ok: true; data: T } | { ok: false; error: string }

const CONNECT_ERROR = 'サーバーに接続できませんでした'

// 一覧は ProjectPage（projects + next_cursor）をそのまま返す:
// next_cursor は「次ページの鍵」としてクライアントの useInfiniteQuery が使う不透明な値
export async function searchProjects(
  params: ProjectsListParams,
): Promise<Result<TsunaguWorksProjectPage>> {
  try {
    const res = await projectSearchClient.list(params)
    if (res.status !== 200) {
      return { ok: false, error: res.data.error ?? '検索に失敗しました' }
    }
    return { ok: true, data: res.data }
  } catch {
    return { ok: false, error: CONNECT_ERROR }
  }
}

// 未公開（draft / closed）と不存在は API がどちらも 404 を返す（存在自体を漏らさない）
export async function getPublishedProject(
  id: number,
): Promise<Result<TsunaguWorksProject>> {
  try {
    const res = await projectSearchClient.get(id)
    if (res.status !== 200) {
      return { ok: false, error: res.data.error ?? '案件が見つかりません' }
    }
    return { ok: true, data: res.data }
  } catch {
    return { ok: false, error: CONNECT_ERROR }
  }
}
