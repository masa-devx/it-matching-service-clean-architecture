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

type ProjectListResponse = {
  projects: Project[]
  limit: number
  offset: number
}

// 公開中の案件一覧。RSC から Go を直接読む（読み取りは BFF を経由しない）
export async function getPublishedProjects(
  limit = 20,
): Promise<Project[] | null> {
  const token = await getTokenCookie()
  if (!token) {
    return null
  }

  const res = await apiGet<ProjectListResponse>(
    `/projects?limit=${limit}`,
    token,
  )
  if (res.error) {
    return null
  }
  return res.data.projects
}
