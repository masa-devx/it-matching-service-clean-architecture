import { NextResponse } from 'next/server'
import { deleteTokenCookie } from '@/lib/authCookie'

export async function POST() {
  await deleteTokenCookie()
  return NextResponse.json({ ok: true }, { status: 200 })
}
