import 'server-only'

import { apiGet } from './api'
import { getTokenCookie } from './authCookie'

// 稼働報告の状態。api/work_report_status.go と対応させる
export const WORK_REPORT_STATUSES = [
  'submitted',
  'approved',
  'rejected',
] as const

export type WorkReportStatus = (typeof WORK_REPORT_STATUSES)[number]

// Go の GET /contracts/{id}/work-reports と同じ形。
// week_start は日付だけの文字列（"2026-08-03"）。時刻もタイムゾーンも意味を持たないため、
// api 側で日付文字列に整形して返している
export type WorkReport = {
  id: number
  contract_id: number
  week_start: string
  hours: number
  summary: string
  status: WorkReportStatus
  // 差し戻しの理由。承認時・未確認時は空文字
  review_note: string
  submitted_at: string
  reviewed_at: string | null
}

export type WorkReportListResult = {
  work_reports: WorkReport[]
  total: number
}

// 契約に紐づく稼働報告の一覧。企業・人材の双方が同じものを見る（共有された記録のため）。
//
// ページネーションは無い。週ごとに1件しか存在せず（UNIQUE制約）、契約期間も
// せいぜい数十週なので、全件返すほうが画面側で週の連続性を扱いやすい
export async function getWorkReports(
  contractId: number,
): Promise<WorkReportListResult | null> {
  const token = await getTokenCookie()
  if (!token) {
    return null
  }

  const res = await apiGet<WorkReportListResult>(
    `/contracts/${contractId}/work-reports`,
    token,
  )
  if (res.error) {
    return null
  }
  return res.data
}
