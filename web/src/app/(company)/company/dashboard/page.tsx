import Link from 'next/link'
import { redirect } from 'next/navigation'

import { LogoutButton } from '@/features/auth/components/client/LogoutButton'
import { meCompany } from '@/external/handler/auth'

export default async function Page() {
  const me = await meCompany()
  if (!me) {
    redirect('/company/login')
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">企業ダッシュボード</h1>
        <LogoutButton role="company" />
      </div>

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

      <Link
        href="/company/projects/new"
        className="inline-block rounded bg-blue-600 px-4 py-2 text-white"
      >
        案件を作成する
      </Link>
    </div>
  )
}
