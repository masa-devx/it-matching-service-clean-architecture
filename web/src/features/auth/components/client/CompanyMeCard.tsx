'use client'

import { useQuery } from '@tanstack/react-query'

import { companyMeQuery } from '../../queries/companyMe'

// RSC で prefetch 済みのキャッシュを引き継ぐクライアント側。
// 初回描画時点でデータがあるためローディング表示は不要（null はガード済みの想定外のみ）
export function CompanyMeCard() {
  const { data: me } = useQuery(companyMeQuery)

  if (!me) {
    return null
  }

  return (
    <dl className="space-y-2 rounded border p-4">
      <div>
        <dt className="text-sm text-gray-500">会社名</dt>
        <dd className="font-medium">{me.name}</dd>
      </div>
      <div>
        <dt className="text-sm text-gray-500">メールアドレス</dt>
        <dd className="font-medium">{me.email}</dd>
      </div>
      {me.location && (
        <div>
          <dt className="text-sm text-gray-500">所在地</dt>
          <dd className="font-medium">{me.location}</dd>
        </div>
      )}
    </dl>
  )
}
