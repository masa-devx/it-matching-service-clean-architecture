import { redirect } from 'next/navigation'

import { AppHeader } from '@/components/AppHeader'
import { LogoutButton } from '@/features/auth/components/client/LogoutButton'
import { meCompany } from '@/external/handler/auth'

// (company) は company ロール必須。判定は Go の me を単一の真実とする
// （talent のトークンは Go 側のロールMWが 403 にするため、ここでも me が null になる）。
// グループ配下の全ページがこの1枚で保護され、アプリシェルもここで合成する
export default async function CompanyLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const me = await meCompany()
  if (!me) {
    redirect('/company/login')
  }

  return (
    <div className="flex min-h-full flex-col">
      <AppHeader role="company" displayName={me.name}>
        <LogoutButton role="company" />
      </AppHeader>
      <main className="mx-auto w-full max-w-5xl flex-1 p-6">{children}</main>
    </div>
  )
}
