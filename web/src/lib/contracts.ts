import 'server-only'

import { apiGet } from './api'
import { getTokenCookie } from './authCookie'

// 契約の状態。api/contract_status.go の CHECK 制約と対応させる
export const CONTRACT_STATUSES = [
  'active',
  'working',
  'reviewing',
  'completed',
  'cancelled',
] as const

export type ContractStatus = (typeof CONTRACT_STATUSES)[number]

// Go の GET /me/contracts と同じ形。
//
// 企業・人材のどちらから見ても同じ形が返る（api 側で視点別に分けていない）。
// 契約は「合意した内容と、その進み具合」を双方が同じように見るため。
// 相手が誰かは、企業なら talent_name、人材なら company_name を使う
export type Contract = {
  id: number
  status: ContractStatus
  // 承諾時点でコピーされた条件。案件を編集してもここは変わらない（#103）
  title: string
  hourly_rate: number
  hours_per_week: number
  remote_ok: boolean
  company_name: string
  talent_name: string
  project_id: number
  started_at: string | null
  completed_at: string | null
  created_at: string
  work_report_count: number
  // まだ企業が確認していない報告の数。企業にとっての「次にやること」を示す
  pending_report_count: number
}

export type ContractListResult = {
  contracts: Contract[]
  total: number
  limit: number
  offset: number
}

export const CONTRACTS_PER_PAGE = 20

export type ContractSearchParams = {
  status?: string
  page?: string
}

export function currentContractPage(params: ContractSearchParams): number {
  const page = Number(params.page)
  return Number.isInteger(page) && page > 0 ? page : 1
}

// 自分が当事者の契約一覧。RSC から Go を直接読む（読み取りは BFF を経由しない）
export async function getMyContracts(
  params: ContractSearchParams = {},
): Promise<ContractListResult | null> {
  const token = await getTokenCookie()
  if (!token) {
    return null
  }

  const query = new URLSearchParams()
  if (params.status) {
    query.set('status', params.status)
  }
  query.set('limit', String(CONTRACTS_PER_PAGE))
  query.set(
    'offset',
    String((currentContractPage(params) - 1) * CONTRACTS_PER_PAGE),
  )

  const res = await apiGet<ContractListResult>(`/me/contracts?${query}`, token)
  if (res.error) {
    return null
  }
  return res.data
}

// 契約の詳細。当事者でない契約・存在しないIDは API が404を返すため null になる
export async function getMyContract(id: number): Promise<Contract | null> {
  const token = await getTokenCookie()
  if (!token) {
    return null
  }

  const res = await apiGet<Contract>(`/me/contracts/${id}`, token)
  if (res.error) {
    return null
  }
  return res.data
}

// ダッシュボードで「対応が必要な件数」を出すための集計。
//
// 一覧APIを1ページ分だけ引いて数えている。件数の上限（20件）を超えると
// 正確でなくなるが、ダッシュボードは「対応が必要かどうか」の気づきを与えるのが目的で、
// 正確な件数は契約一覧で確認できるため許容している。
// 契約数が増えて実害が出たら、API側に集計を足す（#96 → #103 と同じ流れ）
export async function countPendingWorkReports(): Promise<number> {
  const result = await getMyContracts()
  if (!result) {
    return 0
  }
  return result.contracts.reduce(
    (sum, contract) => sum + contract.pending_report_count,
    0,
  )
}

// 稼働中の契約数。人材・企業のどちらのダッシュボードでも「今の仕事」を示す
export async function countActiveContracts(): Promise<number> {
  const result = await getMyContracts({ status: 'working' })
  return result?.total ?? 0
}
