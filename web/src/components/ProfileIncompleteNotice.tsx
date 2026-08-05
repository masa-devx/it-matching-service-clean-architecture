import Link from 'next/link'
import { Button } from '@/components/ui/button'

// プロフィール未登録の案内。必須（企業=案件掲載の前提）か任意かで見せ方を変える
export function ProfileIncompleteNotice({
  message,
  href,
  required = false,
}: {
  message: string
  href: string
  required?: boolean
}) {
  return (
    <div
      className={`flex flex-col gap-3 rounded-lg border p-4 sm:flex-row sm:items-center sm:justify-between ${
        required ? 'border-destructive/40 bg-destructive/5' : 'bg-card'
      }`}
    >
      <p className="text-sm">{message}</p>
      <Button
        asChild
        variant={required ? 'default' : 'outline'}
        className="h-11 shrink-0"
      >
        <Link href={href}>プロフィールを登録する</Link>
      </Button>
    </div>
  )
}
