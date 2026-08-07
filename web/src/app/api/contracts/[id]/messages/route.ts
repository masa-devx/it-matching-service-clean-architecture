import { NextResponse } from 'next/server'
import { apiPost } from '@/lib/api'
import { getTokenCookie } from '@/lib/authCookie'

// メッセージの送信を Go へ中継する。
// マスキングは Go 側が保存時に行う（画面では加工しない——
// 表示のたびに計算すると、正規表現を改善したとき過去の見え方まで変わるため）
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
    `/contracts/${contractId}/messages`,
    { body: body?.body },
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
