'use server'

import type { TsunaguWorksCompanyMe } from '@repo/api-client/company/generated/models'

import { meCompany } from '@/external/handler/auth'

// TanStack Query の queryFn 用の読み取りアクション。
// ブラウザは Go API を直接呼べない（トークンは httpOnly Cookie でクライアントが触れない）ため、
// クライアントからの再フェッチはこの Server Action 経由でサーバーを通す。
// 入力を取らず自分のセッションだけを読むので、誰が呼んでも自分の情報しか返らない
export async function fetchCompanyMe(): Promise<TsunaguWorksCompanyMe | null> {
  return meCompany()
}
