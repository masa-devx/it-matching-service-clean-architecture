import Link from 'next/link'
import { redirect } from 'next/navigation'
import { getCurrentUser } from '@/lib/auth'
import { dashboardPath } from '@/lib/roleRedirect'

// company/(guest) 配下は未ログイン専用（企業向けの認証画面）。
// ログイン済みで開いても意味がないため、自分のロールのダッシュボードへ送る
export default async function CompanyGuestLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const user = await getCurrentUser()
  if (user) {
    redirect(dashboardPath(user.role))
  }

  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-8 px-4 py-10">
      <Link href="/" className="text-2xl font-bold tracking-tight text-primary">
        Tsunagu Works
      </Link>
      {children}
    </div>
  )
}
