import { NextResponse } from 'next/server'
import { apiPut } from '@/lib/api'
import { getTokenCookie } from '@/lib/authCookie'

// 案件の内容を更新する。掲載状態は別ルート（/status）の責務なので、
// ここで status を送っても Go 側が無視する（「保存」と「公開」を分ける設計）
export async function PUT(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const token = await getTokenCookie()
  if (!token) {
    return NextResponse.json({ error: '認証が必要です' }, { status: 401 })
  }

  const { id } = await params
  const projectId = Number(id)
  if (!Number.isInteger(projectId) || projectId <= 0) {
    return NextResponse.json({ error: '案件が見つかりません' }, { status: 400 })
  }

  const body = await req.json()
  const res = await apiPut(`/projects/${projectId}`, body, token)
  if (res.error) {
    return NextResponse.json(
      { error: res.error.message },
      { status: res.error.status || 500 },
    )
  }

  return NextResponse.json({ ok: true }, { status: 200 })
}
