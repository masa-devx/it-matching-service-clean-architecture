import type { ProjectSearchFilters } from './queries/projects'

// URL の searchParams（すべて文字列）→ 検索条件への変換。
// URL が検索状態の一次情報なので、不正値は「条件なし」に落とす（エラーにしない）
export function parseSearchFilters(
  params: Record<string, string | string[] | undefined>,
): ProjectSearchFilters {
  const first = (v: string | string[] | undefined) =>
    Array.isArray(v) ? v[0] : v

  const skills = (first(params.skills) ?? '')
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)

  const remoteOk = first(params.remote_ok) === 'true' ? true : undefined

  const rate = Number(first(params.min_hourly_rate))
  const minHourlyRate = Number.isInteger(rate) && rate > 0 ? rate : undefined

  return { skills, remoteOk, minHourlyRate }
}
