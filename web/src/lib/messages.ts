import 'server-only'

import { apiGet } from './api'
import { getTokenCookie } from './authCookie'
import type { CurrentUser } from './auth'

// Go の GET /contracts/{id}/messages と同じ形。
//
// body には masked_body（伏せ字を入れた文字列）が入る。原文は API が返さないため、
// フロントには到達しない（#117 でレスポンス型に原文の置き場所を作らない設計にした）
export type Message = {
  id: number
  sender_role: CurrentUser['role']
  sender_name: string
  body: string
  // 伏せ字が入ったか。「なぜ消えたのか」を画面で説明するために使う
  masked: boolean
  created_at: string
}

export type MessageListResult = {
  messages: Message[]
  total: number
}

// 契約に紐づく会話。企業・人材の双方が同じものを見る（共有された記録のため）。
//
// ページネーションは無い。1契約のやり取りは限られており、
// 途中で切れると会話の流れが追えなくなる（api 側も全件返す）
export async function getMessages(
  contractId: number,
): Promise<MessageListResult | null> {
  const token = await getTokenCookie()
  if (!token) {
    return null
  }

  const res = await apiGet<MessageListResult>(
    `/contracts/${contractId}/messages`,
    token,
  )
  if (res.error) {
    return null
  }
  return res.data
}
