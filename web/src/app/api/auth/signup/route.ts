import { NextResponse } from 'next/server'
import { apiPost } from '@/lib/api'
import { setTokenCookie } from '@/lib/authCookie'

type LoginResponse = { token: string }

// ブラウザからは Go を直接呼ばず、この Route Handler を経由する。
// トークンはここで httpOnly Cookie に変換し、レスポンスボディには載せない
export async function POST(req: Request) {
  const { email, password, role } = await req.json()

  const signup = await apiPost('/signup', { email, password, role })
  if (signup.error) {
    return NextResponse.json(
      { error: signup.error.message },
      { status: signup.error.status || 500 },
    )
  }

  // 登録後は同じ資格情報でそのまま自動ログイン
  const login = await apiPost<LoginResponse>('/login', { email, password })
  if (login.error) {
    return NextResponse.json(
      { error: '登録は完了しました。ログイン画面からログインしてください' },
      { status: 500 },
    )
  }

  await setTokenCookie(login.data.token)
  return NextResponse.json({ ok: true }, { status: 201 })
}
