import { redirect } from 'next/navigation'
import { getCurrentUser } from '@/lib/auth'
import { dashboardPath } from '@/lib/roleRedirect'

// company/ 配下は企業ユーザー専用。
// 認証は (authenticated)/layout.tsx が担保済みなので、ここではロールだけを見る（多段ガード）。
// URL に権限が表れる構成にすることで、置き場所を間違えた画面が自動的に守られる
export default async function CompanyLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const user = await getCurrentUser()
  if (!user) {
    redirect('/login')
  }
  if (user.role !== 'company') {
    redirect(dashboardPath(user.role))
  }

  return <>{children}</>
}
