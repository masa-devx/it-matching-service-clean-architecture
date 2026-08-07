import { NextResponse } from 'next/server'
import { apiPatch } from '@/lib/api'
import { getTokenCookie } from '@/lib/authCookie'

// 掲載状態の変更を Go へ中継する。遷移が許されるかは Go の遷移表が判定するため、
// ここでは中継に徹する（画面・BFF・APIで判定を三重に書かない）
export async function PATCH(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const token = await getTokenCookie()
  if (!token) {
    return NextResponse.json({ error: '認証が必要です' }, { status: 401 })
  }

  const { id } = await params
  const projectId = Number(id)
  // URLに埋め込む値なので、数値であることをここで確定させる
  if (!Number.isInteger(projectId) || projectId <= 0) {
    return NextResponse.json({ error: '案件が見つかりません' }, { status: 400 })
  }

  const body = await req.json()
  const res = await apiPatch(
    `/projects/${projectId}/status`,
    { status: body?.status },
    token,
  )
  if (res.error) {
    return NextResponse.json(
      { error: res.error.message },
      { status: res.error.status || 500 },
    )
  }

  return NextResponse.json({ ok: true }, { status: 200 })
}
