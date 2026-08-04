import 'server-only'

import { apiGet } from './api'
import { getTokenCookie } from './authCookie'

export type CompanyProfile = {
  name: string
  description: string
  industry: string
  size: string
}

export type TalentProfile = {
  bio: string
  skills: string[]
  years_of_exp: number
  available_hours_per_week: number
  desired_hourly_rate: number
  remote_ok: boolean
}

// Go の GET /me/profile と同じ形。未作成のときは profile が null
export type ProfileResponse =
  | { role: 'company'; profile: CompanyProfile | null }
  | { role: 'talent'; profile: TalentProfile | null }

// RSC から直接 Go を読む（読み取りは BFF 経由にしない）
export async function getProfile(): Promise<ProfileResponse | null> {
  const token = await getTokenCookie()
  if (!token) {
    return null
  }

  const res = await apiGet<ProfileResponse>('/me/profile', token)
  if (res.error) {
    return null
  }
  return res.data
}
