import 'server-only'

import { apiGet } from './api'
import { getTokenCookie } from './authCookie'

// 掲載状態。api/projects.go の CHECK 制約と対応させる
export const PROJECT_STATUSES = ['draft', 'published', 'closed'] as const

export type ProjectStatus = (typeof PROJECT_STATUSES)[number]

// Go の GET /me/projects と同じ形（企業視点の自社案件）。
// 人材向けの Project 型と違い、下書き・募集終了も含み、応募件数を持つ。
// 掲載元の会社は自分自身なので company_name は返らない
export type MyProject = {
  id: number
  title: string
  description: string
  required_skills: string[]
  hourly_rate_min: number
  hourly_rate_max: number
  hours_per_week: number
  remote_ok: boolean
  status: ProjectStatus
  created_at: string
  applications_count: number
  // まだ企業が対応していない応募（applied）。「次にやること」を示す
  pending_count: number
}

export type MyProjectListResult = {
  projects: MyProject[]
  total: number
  limit: number
  offset: number
}

export const MY_PROJECTS_PER_PAGE = 20

export type MyProjectSearchParams = {
  status?: string
  page?: string
}

export function currentMyProjectPage(params: MyProjectSearchParams): number {
  const page = Number(params.page)
  return Number.isInteger(page) && page > 0 ? page : 1
}

// 自社案件の一覧（下書き・募集終了を含む）。応募件数はAPI側が集計して返すため、
// 画面から案件ごとにリクエストする必要はない
export async function getMyProjects(
  params: MyProjectSearchParams = {},
): Promise<MyProjectListResult | null> {
  const token = await getTokenCookie()
  if (!token) {
    return null
  }

  const query = new URLSearchParams()
  if (params.status) {
    query.set('status', params.status)
  }
  query.set('limit', String(MY_PROJECTS_PER_PAGE))
  query.set(
    'offset',
    String((currentMyProjectPage(params) - 1) * MY_PROJECTS_PER_PAGE),
  )

  const res = await apiGet<MyProjectListResult>(`/me/projects?${query}`, token)
  if (res.error) {
    return null
  }
  return res.data
}
