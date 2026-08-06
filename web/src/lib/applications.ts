import 'server-only'

import { apiGet } from './api'
import { getTokenCookie } from './authCookie'

// Go の GET /me/applications と同じ形（人材視点の応募）
export type TalentApplication = {
  id: number
  project_id: number
  project_title: string
  company_name: string
  status: ApplicationStatus
  message: string
  created_at: string
}

// 状態の一覧は api/application_status.go と対応させる。
// 画面はこの型に無い状態を受け取らない前提で書く
export const APPLICATION_STATUSES = [
  'applied',
  'offered',
  'accepted',
  'rejected',
  'withdrawn',
  'declined',
] as const

export type ApplicationStatus = (typeof APPLICATION_STATUSES)[number]

export type TalentApplicationListResult = {
  applications: TalentApplication[]
  total: number
  limit: number
  offset: number
}

export const APPLICATIONS_PER_PAGE = 20

export type ApplicationSearchParams = {
  status?: string
  page?: string
}

export function currentApplicationPage(
  params: ApplicationSearchParams,
): number {
  const page = Number(params.page)
  return Number.isInteger(page) && page > 0 ? page : 1
}

// 自分の応募履歴。RSC から Go を直接読む（読み取りは BFF を経由しない）
export async function getMyApplications(
  params: ApplicationSearchParams = {},
  perPage = APPLICATIONS_PER_PAGE,
): Promise<TalentApplicationListResult | null> {
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

  const res = await apiGet<TalentApplicationListResult>(
    `/me/applications?${query}`,
    token,
  )
  if (res.error) {
    return null
  }
  return res.data
}

// 案件詳細で「すでに応募済みか」を判定するための上限。
// API に project_id の絞り込みが無いため、履歴を引いて突き合わせている。
// 応募数がこれを超えると取りこぼすが、その場合も API 側が二重応募を 409 で拒否する
// （画面の出し分けはUX、正しさはAPIの責務）
const APPLIED_LOOKUP_LIMIT = 100

// 指定案件への自分の応募。未応募なら null
export async function findMyApplicationForProject(
  projectId: number,
): Promise<TalentApplication | null> {
  const result = await getMyApplications({}, APPLIED_LOOKUP_LIMIT)
  if (!result) {
    return null
  }
  return result.applications.find((a) => a.project_id === projectId) ?? null
}
