import 'server-only'

import { apiGet } from './api'
import { getTokenCookie } from './authCookie'

export type Project = {
  id: number
  company_id: number
  company_name: string
  title: string
  description: string
  required_skills: string[]
  hourly_rate_min: number
  hourly_rate_max: number
  hours_per_week: number
  remote_ok: boolean
  status: string
  created_at: string
}

export type ProjectListResult = {
  projects: Project[]
  total: number
  limit: number
  offset: number
}

// 一覧の検索条件。画面のクエリ文字列とAPIのクエリ文字列の橋渡しをする
export type ProjectSearchParams = {
  skills?: string
  rate_min?: string
  rate_max?: string
  hours_max?: string
  remote_ok?: string
  q?: string
  page?: string
}

export const PROJECTS_PER_PAGE = 20

// 画面のURLクエリを Go API のクエリ文字列に変換する。
// 空文字は送らない（APIが「指定なし」と解釈できるようにする）
export function toApiQuery(
  params: ProjectSearchParams,
  perPage = PROJECTS_PER_PAGE,
): string {
  const query = new URLSearchParams()
  for (const key of [
    'skills',
    'rate_min',
    'rate_max',
    'hours_max',
    'remote_ok',
    'q',
  ] as const) {
    const value = params[key]
    if (value) {
      query.set(key, value)
    }
  }

  query.set('limit', String(perPage))
  query.set('offset', String((currentPage(params) - 1) * perPage))
  return query.toString()
}

// ページ番号は1始まり。不正な値は1ページ目に丸める（一覧が壊れるより穏当）
export function currentPage(params: ProjectSearchParams): number {
  const page = Number(params.page)
  return Number.isInteger(page) && page > 0 ? page : 1
}

// 公開中の案件一覧。RSC から Go を直接読む（読み取りは BFF を経由しない）
export async function searchProjects(
  params: ProjectSearchParams = {},
): Promise<ProjectListResult | null> {
  const token = await getTokenCookie()
  if (!token) {
    return null
  }

  const res = await apiGet<ProjectListResult>(
    `/projects?${toApiQuery(params)}`,
    token,
  )
  if (res.error) {
    return null
  }
  return res.data
}

// ダッシュボードの件数表示など、一覧の中身が不要な場面で使う
export async function countPublishedProjects(): Promise<number> {
  const result = await searchProjects({})
  return result?.total ?? 0
}
