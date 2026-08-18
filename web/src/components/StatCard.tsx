import Link from 'next/link'

import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils'

// ダッシュボードの統計カード（数値 + ラベル + 遷移先）。
// highlight は「今やるべきことがある」の強調（判断は呼び出し側・この部品は意味を知らない）
export function StatCard({
  label,
  value,
  href,
  highlight = false,
}: {
  label: string
  value: number
  href: string
  highlight?: boolean
}) {
  return (
    <Link href={href} className="block">
      <Card
        className={cn(
          'transition-shadow hover:shadow-md',
          highlight && 'border-primary bg-primary/5',
        )}
      >
        <CardContent className="space-y-1">
          <p className="text-3xl font-bold">{value}</p>
          <p className="text-sm text-muted-foreground">{label}</p>
        </CardContent>
      </Card>
    </Link>
  )
}
