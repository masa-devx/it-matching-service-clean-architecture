import { redirect } from 'next/navigation'
import { getCurrentUser } from '@/lib/auth'
import { dashboardPath } from '@/lib/roleRedirect'
import { AppHeader } from '@/components/AppHeader'

// talent/(authenticated) 配下は「ログイン済みの人材ユーザー」専用。
// 認証（誰か）とロール（何をしてよいか）を2段で確認し、配下の全ページを一括で守る
export default async function TalentAuthenticatedLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const user = await getCurrentUser()
  if (!user) {
    redirect('/login')
  }
  // 企業ユーザーが人材向けURLに来たら、自分のダッシュボードへ送り返す
  if (user.role !== 'talent') {
    redirect(dashboardPath(user.role))
  }

  return (
    <div className="flex flex-1 flex-col">
      <AppHeader user={user} />
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8">
        {children}
      </main>
    </div>
  )
}
