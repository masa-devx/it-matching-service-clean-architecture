import { NextResponse } from 'next/server'
import { apiPatch } from '@/lib/api'
import { getTokenCookie } from '@/lib/authCookie'

// 稼働報告の承認・差し戻しを Go へ中継する。
// 差し戻し理由の必須チェックは Go 側が行う（画面・BFF・APIで判定を三重に書かない）
export async function PATCH(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const token = await getTokenCookie()
  if (!token) {
    return NextResponse.json({ error: '認証が必要です' }, { status: 401 })
  }

  const { id } = await params
  const workReportId = Number(id)
  // URLに埋め込む値なので、数値であることをここで確定させる
  if (!Number.isInteger(workReportId) || workReportId <= 0) {
    return NextResponse.json(
      { error: '稼働報告が見つかりません' },
      { status: 400 },
    )
  }

  const body = await req.json()
  const res = await apiPatch(
    `/work-reports/${workReportId}/status`,
    { status: body?.status, review_note: body?.review_note },
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
