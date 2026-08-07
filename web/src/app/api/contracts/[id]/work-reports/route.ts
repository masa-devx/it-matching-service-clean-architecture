import { NextResponse } from 'next/server'
import { apiPost } from '@/lib/api'
import { getTokenCookie } from '@/lib/authCookie'

// 稼働報告の新規提出を Go へ中継する。
// week_start は週内のどの日でもよく、丸めは Go 側が date_trunc で行う
export async function POST(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const token = await getTokenCookie()
  if (!token) {
    return NextResponse.json({ error: '認証が必要です' }, { status: 401 })
  }

  const { id } = await params
  const contractId = Number(id)
  if (!Number.isInteger(contractId) || contractId <= 0) {
    return NextResponse.json({ error: '契約が見つかりません' }, { status: 400 })
  }

  const body = await req.json()
  const res = await apiPost(
    `/contracts/${contractId}/work-reports`,
    body,
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
