import Link from 'next/link'
import type { CurrentUser } from '@/lib/auth'
import { navItemsByRole, roleLabel } from '@/lib/nav'
import { dashboardPath } from '@/lib/roleRedirect'
import { LogoutButton } from '@/components/LogoutButton'
import { NavLinks } from '@/components/NavLinks'
import { MobileNav } from '@/components/MobileNav'
import { Badge } from '@/components/ui/badge'

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
          {/* 横幅が足りないモバイルでは Sheet に格納する */}
          <div className="hidden md:block">
            <NavLinks items={navItemsByRole[user.role]} />
          </div>
        </div>
        <div className="flex items-center gap-4">
          <span className="hidden text-sm text-muted-foreground sm:inline">
            {user.email}
          </span>
          {/* ロールは「今どちらのモードか」を示す情報なので常に見せる */}
          <Badge variant="secondary">{roleLabel[user.role]}</Badge>
          <LogoutButton />
          <MobileNav items={navItemsByRole[user.role]} />
        </div>
      </div>
    </header>
  )
}
