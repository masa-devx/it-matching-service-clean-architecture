'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { updateApplicationStatus } from '@/lib/applicationClient'
import type { ApplicationStatus } from '@/lib/applications'
import { Button } from '@/components/ui/button'
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

type Action = {
  status: 'accepted' | 'declined' | 'withdrawn'
  label: string
  // 完了トーストの文言。ボタンの文言から機械的に作ると不自然な日本語になるため個別に持つ
  doneLabel: string
  confirm: { title: string; description: string }
  variant?: 'default' | 'outline'
}

// 状態ごとに「人材が実行できる操作」を並べる。
// api/application_status.go の遷移表のうち roleTalent の行と対応させること
// （ここに無い操作は画面から実行できないが、正しさの担保は API 側の 409）
const actionsByStatus: Partial<Record<ApplicationStatus, Action[]>> = {
  applied: [
    {
      status: 'withdrawn',
      label: '応募を取り下げる',
      doneLabel: '応募を取り下げ',
      variant: 'outline',
      confirm: {
        title: '応募を取り下げますか？',
        description:
          '取り下げると選考は終了し、同じ案件に再度応募することはできません。',
      },
    },
  ],
  offered: [
    {
      status: 'accepted',
      label: '承諾する',
      doneLabel: 'オファーを承諾し',
      confirm: {
        title: 'オファーを承諾しますか？',
        description:
          '承諾すると契約が成立します。条件をよく確認してから実行してください。',
      },
    },
    {
      status: 'declined',
      label: '辞退する',
      doneLabel: 'オファーを辞退し',
      variant: 'outline',
      confirm: {
        title: 'オファーを辞退しますか？',
        description: '辞退すると選考は終了し、取り消すことはできません。',
      },
    },
  ],
}

export function ApplicationActions({
  applicationId,
  status,
}: {
  applicationId: number
  status: ApplicationStatus
}) {
  const router = useRouter()
  const [pending, startTransition] = useTransition()
  const [sending, setSending] = useState(false)

  const actions = actionsByStatus[status]
  if (!actions) {
    return null
  }

  async function run(next: string, doneLabel: string) {
    setSending(true)
    const result = await updateApplicationStatus(applicationId, next)
    setSending(false)

    if (!result.ok) {
      toast.error(result.error)
      // 他の操作で状態が変わっていた場合（409）は、画面を最新に合わせる
      startTransition(() => router.refresh())
      return
    }

    toast.success(`${doneLabel}ました`)
    startTransition(() => router.refresh())
  }

  const disabled = sending || pending

  return (
    <div className="flex flex-wrap gap-2">
      {actions.map((action) => (
        <AlertDialog key={action.status}>
          <AlertDialogTrigger asChild>
            <Button
              variant={action.variant ?? 'default'}
              className="h-11"
              disabled={disabled}
            >
              {action.label}
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{action.confirm.title}</AlertDialogTitle>
              <AlertDialogDescription>
                {action.confirm.description}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>やめる</AlertDialogCancel>
              <AlertDialogAction
                disabled={disabled}
                onClick={() => run(action.status, action.doneLabel)}
              >
                実行する
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      ))}
    </div>
  )
}
