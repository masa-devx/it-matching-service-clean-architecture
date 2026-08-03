import 'server-only'

import { apiGet } from './api'
import { getTokenCookie } from './authCookie'

export type CurrentUser = {
  id: number
  email: string
  role: 'company' | 'talent'
}

// ログイン中のユーザーを返す。未ログイン・トークン不正/期限切れは null。
// Cookieの有無だけで判定せず Go の GET /me に検証させる
// （署名・exp・alg のチェックは API 側の実装(#5)を単一の真実として使う）
export async function getCurrentUser(): Promise<CurrentUser | null> {
  const token = await getTokenCookie()
  if (!token) {
    return null
  }

  const me = await apiGet<CurrentUser>('/me', token)
  if (me.error) {
    return null
  }
  return me.data
}
