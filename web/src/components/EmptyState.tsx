import type { LucideIcon } from 'lucide-react'

// 空状態の統一フォーマット（アイコン + タイトル + 説明 + 次の行動）。
// 「何も出ない＝壊れてる？」を防ぎ、次にすべきことを必ず示す。
// アクションは children スロット（この部品はどの画面のことも知らない）
export function EmptyState({
  icon: Icon,
  title,
  description,
  children,
}: {
  icon: LucideIcon
  title: string
  description?: string
  children?: React.ReactNode
}) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-lg border border-dashed p-10 text-center">
      <Icon className="size-8 text-muted-foreground" aria-hidden="true" />
      <div className="space-y-1">
        <p className="font-medium">{title}</p>
        {description && (
          <p className="text-sm text-muted-foreground">{description}</p>
        )}
      </div>
      {children}
    </div>
  )
}
