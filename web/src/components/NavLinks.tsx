'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import type { NavItem } from '@/lib/nav'
import { cn } from '@/lib/utils'

// 現在地の判定にパスが必要なため、ここだけクライアントコンポーネントにする
// （ヘッダー全体をクライアント化せず、葉に押し下げる）
export function NavLinks({ items }: { items: NavItem[] }) {
  const pathname = usePathname()

  return (
    <nav className="flex items-center gap-4 text-sm">
      {items.map((item) => {
        const isActive = pathname.startsWith(item.href)
        return (
          <Link
            key={item.href}
            href={item.href}
            aria-current={isActive ? 'page' : undefined}
            className={cn(
              'transition-colors hover:text-foreground',
              isActive
                ? 'font-medium text-foreground'
                : 'text-muted-foreground',
            )}
          >
            {item.label}
          </Link>
        )
      })}
    </nav>
  )
}
