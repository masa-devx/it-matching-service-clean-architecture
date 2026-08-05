'use client'

import { Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'

// 送信ボタンの共通実装。送信中は「無効化＋スピナー＋文言変更」の3点セットで
// 状態を伝える（文言だけだと変化に気づきにくい）
export function SubmitButton({
  isSubmitting,
  children,
  submittingLabel,
  className,
}: {
  isSubmitting: boolean
  children: React.ReactNode
  submittingLabel: string
  className?: string
}) {
  return (
    <Button
      type="submit"
      disabled={isSubmitting}
      className={className ?? 'h-11 self-start'}
    >
      {isSubmitting && (
        <Loader2 className="size-4 animate-spin" aria-hidden="true" />
      )}
      {isSubmitting ? submittingLabel : children}
    </Button>
  )
}
