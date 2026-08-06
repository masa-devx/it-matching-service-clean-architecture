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
  status: 'offered' | 'rejected'
  label: string
  doneLabel: string
  confirm: { title: string; description: string }
  variant?: 'default' | 'outline'
}

// 状態ごとに「企業が実行できる操作」を並べる。
// api/application_status.go の遷移表のうち roleCompany の行と対応させること。
// 企業が実行できるのは applied からの2つだけ（offered の先＝成立させるのは人材の操作）
const actionsByStatus: Partial<Record<ApplicationStatus, Action[]>> = {
  applied: [
    {
      status: 'offered',
      label: 'オファーする',
      doneLabel: 'オファーし',
      confirm: {
        title: 'この応募者にオファーしますか？',
        description:
          '相手に通知され、承諾されると契約が成立します。オファーは取り消せません。',
      },
    },
    {
      status: 'rejected',
      label: '見送る',
      doneLabel: '見送りに',
      variant: 'outline',
      confirm: {
        title: 'この応募を見送りますか？',
        description:
          '相手に見送りとして伝わります。取り消して選考を再開することはできません。',
      },
    },
  ],
}

export function ApplicantActions({
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
      // 人材が先に取り下げていた場合（409）は、画面を最新に合わせる
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
