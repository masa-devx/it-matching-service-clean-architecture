import 'server-only'

import { apiGet } from './api'
import { getTokenCookie } from './authCookie'
import type { ApplicationStatus, ApplicationSearchParams } from './applications'
import { currentApplicationPage, APPLICATIONS_PER_PAGE } from './applications'

// Go の GET /projects/{id}/applications と同じ形（企業視点の応募）。
// 人材視点（TalentApplication）と別の型にする：見たいものが違うため
export type CompanyApplication = {
  id: number
  talent_id: number
  display_name: string
  skills: string[]
  years_of_exp: number
  status: ApplicationStatus
  message: string
  created_at: string
  // 企業・人材それぞれの意思表示の時刻（未実施なら null）。
  // 両方埋まっている＝双方が意思表示済み＝ダブルオプトインが成立している
  offered_at: string | null
  answered_at: string | null
}

export type CompanyApplicationListResult = {
  applications: CompanyApplication[]
  total: number
  limit: number
  offset: number
}

// 自社案件への応募一覧。他社の案件IDなら API が 404 を返すため null になる
// （呼び出し側で notFound() に変換する）
export async function getProjectApplications(
  projectId: number,
  params: ApplicationSearchParams = {},
  perPage = APPLICATIONS_PER_PAGE,
): Promise<CompanyApplicationListResult | null> {
  const token = await getTokenCookie()
  if (!token) {
    return null
  }

  const query = new URLSearchParams()
  if (params.status) {
    query.set('status', params.status)
  }
  query.set('limit', String(perPage))
  query.set('offset', String((currentApplicationPage(params) - 1) * perPage))

  const res = await apiGet<CompanyApplicationListResult>(
    `/projects/${projectId}/applications?${query}`,
    token,
  )
  if (res.error) {
    return null
  }
  return res.data
}
