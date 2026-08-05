import Link from 'next/link'
import { Building2, User } from 'lucide-react'
import type { CurrentUser } from '@/lib/auth'
import { dashboardPath } from '@/lib/roleRedirect'
import { Button } from '@/components/ui/button'

// 二面市場（企業⇔人材）のサービスでは、来訪者がどちら側か分からない。
// 「あなたはどちらですか」を最初に問い、経路を分けるのが Path selection パターン
const entries = [
  {
    role: 'company' as const,
    icon: Building2,
    label: '企業の方',
    description: '案件を掲載して人材を探す',
  },
  {
    role: 'talent' as const,
    icon: User,
    label: '人材の方',
    description: '案件を探して応募する',
  },
]

export function Hero({ user }: { user: CurrentUser | null }) {
  return (
    <section className="flex flex-col items-center gap-8 px-4 py-16 text-center sm:py-24">
      <div className="flex max-w-2xl flex-col gap-4">
        <h1 className="text-3xl font-bold leading-tight tracking-tight sm:text-4xl">
          知らない者同士が、
          <br className="sm:hidden" />
          安心して取引できる場所
        </h1>
        <p className="text-base text-muted-foreground sm:text-lg">
          企業とIT人材（副業・フリーランス）をつなぐビジネスマッチング。
          エスクロー決済・連絡先マスキング・公平なレビューで、
          「お金」と「信頼」の不安を仕組みで解決します。
        </p>
      </div>

      {user ? (
        <Button asChild className="h-11">
          <Link href={dashboardPath(user.role)}>ダッシュボードへ</Link>
        </Button>
      ) : (
        <div className="grid w-full max-w-2xl gap-4 sm:grid-cols-2">
          {entries.map(({ role, icon: Icon, label, description }) => (
            <Link
              key={role}
              href={`/${role}/signup`}
              className="flex flex-col items-center gap-3 rounded-lg border bg-card p-6 transition-colors hover:border-primary hover:bg-primary/5"
            >
              <Icon className="size-8 text-primary" aria-hidden="true" />
              <div className="flex flex-col gap-1">
                <span className="font-bold">{label}</span>
                <span className="text-sm text-muted-foreground">
                  {description}
                </span>
              </div>
              <span className="text-sm font-medium text-primary">
                無料で始める →
              </span>
            </Link>
          ))}
        </div>
      )}
    </section>
  )
}
