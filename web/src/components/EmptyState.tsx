import type { LucideIcon } from 'lucide-react'

// 空状態の共通表現。「何も無い」だけで終わらせず、次の行動を示すことを型で強制する
// （action を省略できるが、原則は導線を置く）
export function EmptyState({
  icon: Icon,
  title,
  description,
  action,
}: {
  icon: LucideIcon
  title: string
  description?: string
  action?: React.ReactNode
}) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-lg border bg-card px-6 py-12 text-center">
      {/* アイコンは装飾なのでスクリーンリーダーからは隠す（意味は title が担う） */}
      <Icon className="size-8 text-muted-foreground" aria-hidden="true" />
      <div className="flex flex-col gap-1">
        <p className="font-medium">{title}</p>
        {description && (
          <p className="text-sm text-muted-foreground">{description}</p>
        )}
      </div>
      {action}
    </div>
  )
}
