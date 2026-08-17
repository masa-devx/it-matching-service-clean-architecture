import { redirect } from 'next/navigation'

import { LogoutButton } from '@/features/auth/components/client/LogoutButton'
import { meTalent } from '@/external/handler/auth'

export default async function Page() {
  const me = await meTalent()
  if (!me) {
    redirect('/talent/login')
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">人材ダッシュボード</h1>
        <LogoutButton role="talent" />
      </div>

      <dl className="space-y-2 rounded border p-4">
        <div>
          <dt className="text-sm text-gray-500">表示名</dt>
          <dd className="font-medium">{me.display_name}</dd>
        </div>
        <div>
          <dt className="text-sm text-gray-500">メールアドレス</dt>
          <dd className="font-medium">{me.email}</dd>
        </div>
        <div>
          <dt className="text-sm text-gray-500">スキル</dt>
          <dd className="font-medium">
            {me.skills.length > 0 ? me.skills.join(' / ') : '未登録'}
          </dd>
        </div>
      </dl>
    </div>
  )
}
