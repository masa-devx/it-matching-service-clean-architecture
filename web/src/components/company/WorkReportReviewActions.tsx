'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { reviewWorkReport } from '@/lib/contractClient'
import type { WorkReport } from '@/lib/workReports'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'

// 稼働報告の承認・差し戻し（企業）。
//
// 操作できるのは「確認待ち（submitted）」の報告だけ。承認済みは覆せず、
// 差し戻し中は人材が修正している最中なので、企業からは触れない
// （api/work_report_status.go の遷移表と対応させること）
export function WorkReportReviewActions({ report }: { report: WorkReport }) {
  const router = useRouter()
  const [pending, startTransition] = useTransition()
  const [sending, setSending] = useState(false)
  const [reviewNote, setReviewNote] = useState('')

  if (report.status !== 'submitted') {
    return null
  }

  const disabled = sending || pending

  async function run(status: 'approved' | 'rejected', note: string) {
    setSending(true)
    const result = await reviewWorkReport(report.id, status, note)
    setSending(false)

    if (!result.ok) {
      toast.error(result.error)
      // 人材が先に再提出していた場合（409）は、画面を最新に合わせる
      startTransition(() => router.refresh())
      return
    }

    toast.success(status === 'approved' ? '承認しました' : '差し戻しました')
    setReviewNote('')
    startTransition(() => router.refresh())
  }

  return (
    <div className="flex flex-wrap gap-2">
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button className="h-11" disabled={disabled}>
            承認する
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>この稼働報告を承認しますか？</AlertDialogTitle>
            <AlertDialogDescription>
              承認すると確定した実績として記録され、あとから取り消すことはできません。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>やめる</AlertDialogCancel>
            <AlertDialogAction
              disabled={disabled}
              onClick={() => run('approved', '')}
            >
              承認する
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* 差し戻しは理由が必須（api 側も400で拒否する）。
          単なる確認ではなく入力を伴うため、ダイアログの中に入力欄を置く */}
      <AlertDialog
        onOpenChange={(open) => {
          // 閉じたら入力を捨てる。前回の理由が残っていると、
          // 別の報告を差し戻すときに誤って送信してしまう
          if (!open) {
            setReviewNote('')
          }
        }}
      >
        <AlertDialogTrigger asChild>
          <Button variant="outline" className="h-11" disabled={disabled}>
            差し戻す
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>この稼働報告を差し戻しますか？</AlertDialogTitle>
            <AlertDialogDescription>
              人材が内容を修正して再提出します。何を直せばよいかが伝わるように書いてください。
            </AlertDialogDescription>
          </AlertDialogHeader>

          <div className="flex flex-col gap-2">
            <Label htmlFor={`review-note-${report.id}`}>差し戻しの理由</Label>
            <Textarea
              id={`review-note-${report.id}`}
              rows={4}
              placeholder="例: 打ち合わせの時間も稼働に含めてください"
              value={reviewNote}
              onChange={(e) => setReviewNote(e.target.value)}
            />
          </div>

          <AlertDialogFooter>
            <AlertDialogCancel>やめる</AlertDialogCancel>
            <AlertDialogAction
              // 理由が空なら押せない。サーバーも400を返すが、
              // 送信してからエラーになるより、押す前に分かるほうが親切
              disabled={disabled || reviewNote.trim() === ''}
              onClick={() => run('rejected', reviewNote.trim())}
            >
              差し戻す
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
