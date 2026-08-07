import { NextResponse } from 'next/server'
import { apiPut } from '@/lib/api'
import { getTokenCookie } from '@/lib/authCookie'

// 差し戻された稼働報告の修正＋再提出を中継する。
// 内容の更新と状態の復帰を1操作で行う設計は Go 側と揃えている
// （稼働報告に下書きの概念が無く、内容を出すこと自体が提出であるため）
export async function PUT(
  req: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const token = await getTokenCookie()
  if (!token) {
    return NextResponse.json({ error: '認証が必要です' }, { status: 401 })
  }

  const { id } = await params
  const workReportId = Number(id)
  if (!Number.isInteger(workReportId) || workReportId <= 0) {
    return NextResponse.json(
      { error: '稼働報告が見つかりません' },
      { status: 400 },
    )
  }

  const body = await req.json()
  const res = await apiPut(`/work-reports/${workReportId}`, body, token)
  if (res.error) {
    return NextResponse.json(
      { error: res.error.message },
      { status: res.error.status || 500 },
    )
  }

  return NextResponse.json({ ok: true }, { status: 200 })
}
