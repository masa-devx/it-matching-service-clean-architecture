import { NextResponse } from 'next/server'
import { apiPost } from '@/lib/api'
import { getTokenCookie } from '@/lib/authCookie'

// レビューの投稿を Go へ中継する。
// 公開の判定（両者が提出したか）は Go 側がトランザクション内で行う——
// ここで判定すると、判定と更新の間に相手が提出した場合に整合が取れなくなる
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
  // URLに埋め込む値なので、数値であることをここで確定させる
  if (!Number.isInteger(contractId) || contractId <= 0) {
    return NextResponse.json({ error: '契約が見つかりません' }, { status: 400 })
  }

  const body = await req.json()
  const res = await apiPost(
    `/contracts/${contractId}/reviews`,
    { rating: body?.rating, comment: body?.comment },
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
