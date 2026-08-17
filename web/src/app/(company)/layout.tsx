import { redirect } from 'next/navigation'

import { meCompany } from '@/external/handler/auth'

// (company) は company ロール必須。判定は Go の me を単一の真実とする
// （talent のトークンは Go 側のロールMWが 403 にするため、ここでも me が null になる）。
// グループ配下の全ページがこの1枚で保護される
export default async function CompanyLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const me = await meCompany()
  if (!me) {
    redirect('/company/login')
  }

  return <main className="mx-auto max-w-2xl p-8">{children}</main>
}
