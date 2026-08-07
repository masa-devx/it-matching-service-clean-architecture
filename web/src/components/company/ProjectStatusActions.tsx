'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { updateProjectStatus } from '@/lib/projectClient'
import type { ProjectStatus } from '@/lib/companyProjects'
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
  status: ProjectStatus
  label: string
  doneLabel: string
  variant?: 'default' | 'outline'
  // 応募者に影響する操作だけ確認を挟む。公開・再募集は取り返しがつくので確認しない
  confirm?: { title: string; description: string }
}

// 状態ごとに実行できる操作を並べる。
// api/project_status.go の projectTransitions と対応させること
// （ここに無い操作は画面から実行できないが、正しさの担保は API 側の 409）
const actionsByStatus: Record<ProjectStatus, Action[]> = {
  draft: [{ status: 'published', label: '公開する', doneLabel: '公開し' }],
  published: [
    {
      status: 'closed',
      label: '募集を終了する',
      doneLabel: '募集を終了し',
      variant: 'outline',
      confirm: {
        title: 'この案件の募集を終了しますか？',
        description:
          '人材の一覧から外れ、新しい応募を受け付けなくなります。選考中の応募はそのまま残ります。あとで再募集できます。',
      },
    },
    {
      status: 'draft',
      label: '非公開に戻す',
      doneLabel: '非公開に',
      variant: 'outline',
      confirm: {
        title: 'この案件を非公開に戻しますか？',
        description:
          '人材の一覧・詳細から見えなくなります。すでに届いている応募は残り、選考は続けられます。',
      },
    },
  ],
  closed: [
    { status: 'published', label: '再募集する', doneLabel: '再募集を開始し' },
  ],
}

export function ProjectStatusActions({
  projectId,
  status,
}: {
  projectId: number
  status: ProjectStatus
}) {
  const router = useRouter()
  const [pending, startTransition] = useTransition()
  const [sending, setSending] = useState(false)

  const actions = actionsByStatus[status]
  const disabled = sending || pending

  async function run(next: ProjectStatus, doneLabel: string) {
    setSending(true)
    const result = await updateProjectStatus(projectId, next)
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

  return (
    <div className="flex flex-wrap gap-2">
      {actions.map((action) =>
        action.confirm ? (
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
        ) : (
          <Button
            key={action.status}
            variant={action.variant ?? 'default'}
            className="h-11"
            disabled={disabled}
            onClick={() => run(action.status, action.doneLabel)}
          >
            {action.label}
          </Button>
        ),
      )}
    </div>
  )
}
