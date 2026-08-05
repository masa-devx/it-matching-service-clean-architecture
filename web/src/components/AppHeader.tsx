import Link from 'next/link'
import type { CurrentUser } from '@/lib/auth'
import { navItemsByRole, roleLabel } from '@/lib/nav'
import { dashboardPath } from '@/lib/roleRedirect'
import { LogoutButton } from '@/components/LogoutButton'
import { NavLinks } from '@/components/NavLinks'

// 企業・人材のどちらのアプリからも使う共通ヘッダー。
// 表示内容（ナビ・ロール表示）はロールから導出する
export function AppHeader({ user }: { user: CurrentUser }) {
  return (
    <header className="border-b bg-card">
      <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between px-4">
        <div className="flex items-center gap-8">
          <Link
            href={dashboardPath(user.role)}
            className="text-lg font-bold text-primary"
          >
            Tsunagu Works
          </Link>
          <NavLinks items={navItemsByRole[user.role]} />
        </div>
        <div className="flex items-center gap-4">
          <span className="hidden text-sm text-muted-foreground sm:inline">
            {user.email}
          </span>
          {/* ロールは「今どちらのモードか」を示す情報なので常に見せる */}
          <span className="rounded-full bg-primary/10 px-2.5 py-1 text-xs font-medium text-primary">
            {roleLabel[user.role]}
          </span>
          <LogoutButton />
        </div>
      </div>
    </header>
  )
}
