'use client'

import { Menu } from 'lucide-react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'

import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { cn } from '@/lib/utils'

// 認証済み画面の共通ヘッダー（アプリシェル）。
// 認証情報は layout が解決して props で渡し、ログアウト等の操作部品は children スロットで
// 注入する＝この部品は features を import しない（landing ヘッダーと同じ「判断済みの結果を渡す」型）
type Role = 'company' | 'talent'

const navByRole: Record<Role, Array<{ href: string; label: string }>> = {
  company: [
    { href: '/company/dashboard', label: 'ダッシュボード' },
    { href: '/company/projects', label: '案件管理' },
  ],
  talent: [
    { href: '/talent/dashboard', label: 'ダッシュボード' },
    { href: '/talent/projects', label: '案件を探す' },
    { href: '/talent/applications', label: '応募一覧' },
  ],
}

export function AppHeader({
  role,
  displayName,
  children,
}: {
  role: Role
  displayName: string
  children?: React.ReactNode
}) {
  const pathname = usePathname()
  const items = navByRole[role]

  const isActive = (href: string) =>
    pathname === href || pathname.startsWith(`${href}/`)

  return (
    <header className="sticky top-0 z-10 border-b bg-card/80 backdrop-blur">
      <div className="mx-auto flex h-14 w-full max-w-5xl items-center gap-4 px-4">
        {/* モバイル: ナビをドロワーに収納 */}
        <Sheet>
          <SheetTrigger asChild>
            <Button
              variant="ghost"
              size="icon-sm"
              className="md:hidden"
              aria-label="メニューを開く"
            >
              <Menu />
            </Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-64">
            <SheetHeader>
              <SheetTitle>Tsunagu Works</SheetTitle>
            </SheetHeader>
            <nav className="flex flex-col gap-1 px-4">
              {items.map((item) => (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    'rounded-md px-3 py-2 text-sm hover:bg-muted',
                    isActive(item.href) && 'bg-muted font-medium',
                  )}
                >
                  {item.label}
                </Link>
              ))}
            </nav>
          </SheetContent>
        </Sheet>

        <Link
          href={`/${role}/dashboard`}
          className="text-lg font-bold text-primary"
        >
          Tsunagu Works
        </Link>

        {/* デスクトップ: 水平ナビ。現在地は下線と色で強調 */}
        <nav className="hidden items-center gap-1 md:flex">
          {items.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              aria-current={isActive(item.href) ? 'page' : undefined}
              className={cn(
                'rounded-md px-3 py-1.5 text-sm text-muted-foreground hover:bg-muted hover:text-foreground',
                isActive(item.href) && 'bg-muted font-medium text-foreground',
              )}
            >
              {item.label}
            </Link>
          ))}
        </nav>

        <div className="ml-auto">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm">
                {displayName}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>
                {role === 'company' ? '企業アカウント' : '人材アカウント'}
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              {/* ログアウト等の操作は layout から注入される（features への依存を作らない） */}
              <div className="px-2 py-1.5">{children}</div>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  )
}
