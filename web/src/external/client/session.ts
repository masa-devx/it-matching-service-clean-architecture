import 'server-only'

import { cookies } from 'next/headers'

// トークンは httpOnly Cookie にだけ置く（ブラウザ JS からは読めない＝XSS でも盗めない）。
// この Cookie を触れるのはサーバー側（Server Actions / RSC）だけ
const TOKEN_COOKIE = 'token'

export async function setSessionToken(token: string): Promise<void> {
  const store = await cookies()
  store.set(TOKEN_COOKIE, token, {
    httpOnly: true, // JS から読めない（XSS 対策の核心）
    sameSite: 'lax', // 外部サイト起点の POST に Cookie を載せない（CSRF 緩和）
    secure: process.env.NODE_ENV === 'production', // 本番は https のみ
    path: '/',
    maxAge: 60 * 60 * 24, // JWT の TTL（24h）と揃える
  })
}

export async function getSessionToken(): Promise<string | undefined> {
  const store = await cookies()
  return store.get(TOKEN_COOKIE)?.value
}

export async function clearSessionToken(): Promise<void> {
  const store = await cookies()
  store.delete(TOKEN_COOKIE)
}
