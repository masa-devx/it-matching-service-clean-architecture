import { NextResponse } from 'next/server'
import { apiPut } from '@/lib/api'
import { getTokenCookie } from '@/lib/authCookie'

// 書き込みは BFF 経由。Cookie のトークンを Bearer に変換して Go へ中継する
export async function PUT(req: Request) {
  const token = await getTokenCookie()
  if (!token) {
    return NextResponse.json({ error: '認証が必要です' }, { status: 401 })
  }

  const body = await req.json()
  const res = await apiPut('/me/profile', body, token)
  if (res.error) {
    return NextResponse.json(
      { error: res.error.message },
      { status: res.error.status || 500 },
    )
  }

  return NextResponse.json({ ok: true }, { status: 200 })
}
