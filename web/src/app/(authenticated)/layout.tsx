import Link from 'next/link'
import { redirect } from 'next/navigation'
import { getCurrentUser } from '@/lib/auth'
import { LogoutButton } from '@/components/LogoutButton'

// (authenticated) 配下は要ログイン。ここで一括ガードするため、各ページに認証チェックを書かない
export default async function MainLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  const user = await getCurrentUser()
  if (!user) {
    redirect('/login')
  }

  return (
    <div className="flex flex-1 flex-col">
      <header className="border-b bg-card">
        <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between px-4">
          <div className="flex items-center gap-8">
            <Link href="/projects" className="text-lg font-bold text-primary">
              Tsunagu Works
            </Link>
            <nav className="flex items-center gap-4 text-sm">
              <Link
                href="/projects"
                className="text-muted-foreground hover:text-foreground"
              >
                案件一覧
              </Link>
            </nav>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-muted-foreground">
              {user.email}
              <span className="ml-2 rounded-full bg-secondary px-2 py-0.5 text-xs">
                {user.role === 'company' ? '企業' : '人材'}
              </span>
            </span>
            <LogoutButton />
          </div>
        </div>
      </header>
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-8">
        {children}
      </main>
    </div>
  )
}
