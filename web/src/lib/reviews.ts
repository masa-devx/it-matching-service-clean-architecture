import 'server-only'

import { apiGet } from './api'
import { getTokenCookie } from './authCookie'
import type { CurrentUser } from './auth'

// Go の GET /contracts/{id}/reviews と同じ形。
//
// 未公開の相手レビューは配列に含まれない（API がそもそも SELECT しない）。
// 「画面で隠す」のではなく「取得しない」ことで、開発者ツールやAPIの直接呼び出しでも
// 読めないようにしている
export type Review = {
  id: number
  reviewer_role: CurrentUser['role']
  rating: number
  comment: string
  submitted_at: string
  published_at: string | null
}

// published / submitted の2つのフラグで画面の3状態を出し分ける。
//
// 配列の中身だけでは判断できない——空配列が「相手が未提出」なのか
// 「自分も未提出」なのか区別がつかないため
export type ReviewListResult = {
  reviews: Review[]
  // 双方のレビューが公開されたか
  published: boolean
  // 自分がすでに提出したか
  submitted: boolean
}

// 契約のレビュー。当事者でない契約は API が404を返すため null になる
export async function getReviews(
  contractId: number,
): Promise<ReviewListResult | null> {
  const token = await getTokenCookie()
  if (!token) {
    return null
  }

  const res = await apiGet<ReviewListResult>(
    `/contracts/${contractId}/reviews`,
    token,
  )
  if (res.error) {
    return null
  }
  return res.data
}
