import Link from 'next/link'
import { redirect } from 'next/navigation'
import { getCurrentUser } from '@/lib/auth'
import { dashboardPath } from '@/lib/roleRedirect'

// (guest) 配下（login / signup）は未ログイン専用。
// ログイン済みで認証画面を開いても意味がないため、ロール別のダッシュボードへ送る
export default async function GuestLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
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
