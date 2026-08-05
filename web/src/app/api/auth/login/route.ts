import { NextResponse } from 'next/server'
import { apiGet, apiPost } from '@/lib/api'
import { setTokenCookie } from '@/lib/authCookie'
import { dashboardPath } from '@/lib/roleRedirect'
import type { CurrentUser } from '@/lib/auth'

type LoginResponse = { token: string }

export async function POST(req: Request) {
  const { email, password } = await req.json()

  const login = await apiPost<LoginResponse>('/login', { email, password })
  if (login.error) {
    return NextResponse.json(
      { error: login.error.message },
      { status: login.error.status || 500 },
    )
  }

  await setTokenCookie(login.data.token)

  // 遷移先はサーバー側で決めてクライアントに渡す。
  // クライアントに role を持たせず「行き先」だけを教えることで、
  // ロール→パスの対応（roleRedirect.ts）を1か所に閉じ込められる
  const me = await apiGet<CurrentUser>('/me', login.data.token)
  const redirectTo = me.error ? '/' : dashboardPath(me.data.role)

  return NextResponse.json({ ok: true, redirectTo }, { status: 200 })
}
