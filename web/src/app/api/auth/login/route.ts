import { NextResponse } from 'next/server'
import { apiPost } from '@/lib/api'
import { setTokenCookie } from '@/lib/authCookie'

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
  return NextResponse.json({ ok: true }, { status: 200 })
}
