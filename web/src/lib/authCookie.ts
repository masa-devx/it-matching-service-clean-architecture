import 'server-only'

import { cookies } from 'next/headers'

const TOKEN_COOKIE = 'token'

// JWT を httpOnly Cookie に保存する。
// httpOnly: クライアントJSから読めない（XSSでトークンを盗まれない）
// sameSite=lax: 他サイト起点のリクエストにCookieが乗らない（CSRF緩和）
export async function setTokenCookie(token: string) {
  const cookieStore = await cookies()
  cookieStore.set(TOKEN_COOKIE, token, {
    httpOnly: true,
    sameSite: 'lax',
    secure: process.env.NODE_ENV === 'production',
    path: '/',
    maxAge: 60 * 60 * 24, // Go API 側の JWT exp（24時間）と揃える
  })
}

// ログアウト＝Cookieの削除。JWTはステートレスなのでサーバー側に無効化処理はなく、
// ブラウザからトークンを取り除くことがログアウトの実体になる
export async function deleteTokenCookie() {
  const cookieStore = await cookies()
  cookieStore.delete(TOKEN_COOKIE)
}

export async function getTokenCookie(): Promise<string | undefined> {
  const cookieStore = await cookies()
  return cookieStore.get(TOKEN_COOKIE)?.value
}
