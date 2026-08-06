import { NextResponse } from 'next/server'
import { apiPost } from '@/lib/api'
import { getTokenCookie } from '@/lib/authCookie'

// 書き込みは BFF 経由。Cookie のトークンを Bearer に変換して Go へ中継する。
// 応募先の案件IDはパスではなくボディで受け、Go 側のパス（/projects/{id}/applications）に組み替える
export async function POST(req: Request) {
  const token = await getTokenCookie()
  if (!token) {
    return NextResponse.json({ error: '認証が必要です' }, { status: 401 })
  }

  const body = await req.json()
  const projectId = Number(body?.project_id)
  // URLに埋め込む値なので、数値であることをここで確定させる
  if (!Number.isInteger(projectId) || projectId <= 0) {
    return NextResponse.json({ error: '案件が見つかりません' }, { status: 400 })
  }

  const res = await apiPost(
    `/projects/${projectId}/applications`,
    { message: body?.message },
    token,
  )
  if (res.error) {
    return NextResponse.json(
      { error: res.error.message },
      { status: res.error.status || 500 },
    )
  }

  return NextResponse.json({ ok: true }, { status: 201 })
}
