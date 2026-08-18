'use client'

import type {
  TsunaguWorksApplication,
  TsunaguWorksApplicationStatus,
} from '@repo/api-client/talent/generated/models'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

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
import { Button } from '@/components/ui/button'
import type { ApplicationAction } from '@/external/handler/application'

import { changeApplicationAction } from '../../actions/status'
import { applicationKeys } from '../../queries/applications'

// 状態ごとに出す操作（actor=talent の遷移表の UI 側の写し・#59 company 側と対）。
// 承諾は合意の成立＝最も不可逆な操作なので、確認文でもその重みを伝える。
// 正しさの保証はサーバー（条件付きUPDATE）。競合時は 409（現在: x）がトーストで返る
const successMessage: Record<ApplicationAction, string> = {
  withdraw: 'への応募を取り下げました',
  accept: 'のオファーを承諾しました',
  decline: 'のオファーを辞退しました',
}

const actionsByStatus: Record<
  TsunaguWorksApplicationStatus,
  Array<{
    action: ApplicationAction
    label: string
    variant: 'default' | 'outline' | 'destructive'
    confirm: { title: string; description: string }
  }>
> = {
  applied: [
    {
      action: 'withdraw',
      label: '取り下げる',
      variant: 'outline',
      confirm: {
        title: 'この応募を取り下げますか？',
        description: '取り下げると同じ案件に再応募できません。',
      },
    },
  ],
  offered: [
    {
      action: 'accept',
      label: '承諾する',
      variant: 'default',
      confirm: {
        title: 'オファーを承諾しますか？',
        description:
          '承諾すると双方の合意が成立します。この操作は取り消せません。',
      },
    },
    {
      action: 'decline',
      label: '辞退する',
      variant: 'destructive',
      confirm: {
        title: 'オファーを辞退しますか？',
        description: '辞退すると取り消せません。',
      },
    },
    {
      action: 'withdraw',
      label: '取り下げる',
      variant: 'outline',
      confirm: {
        title: 'この応募を取り下げますか？',
        description: '取り下げると同じ案件に再応募できません。',
      },
    },
  ],
  accepted: [],
  rejected: [],
  withdrawn: [],
  declined: [],
}

export function ApplicationActions({
  application,
}: {
  application: TsunaguWorksApplication
}) {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: async (action: ApplicationAction) => {
      const result = await changeApplicationAction(application.id, action)
      if (!result.ok) {
        throw new Error(result.error)
      }
      return result.data
    },
    onSuccess: (_updated, action) => {
      toast.success(`「${application.project_title}」${successMessage[action]}`)
      void queryClient.invalidateQueries({ queryKey: applicationKeys.all })
    },
    onError: (error) => {
      toast.error(error.message)
    },
  })

  const actions = actionsByStatus[application.status]
  if (actions.length === 0) {
    return null
  }

  return (
    <div className="flex gap-2">
      {actions.map(({ action, label, variant, confirm }) => (
        <AlertDialog key={action}>
          <AlertDialogTrigger asChild>
            <Button variant={variant} size="sm" disabled={mutation.isPending}>
              {label}
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{confirm.title}</AlertDialogTitle>
              <AlertDialogDescription>
                「{application.project_title}」 — {confirm.description}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>キャンセル</AlertDialogCancel>
              <AlertDialogAction onClick={() => mutation.mutate(action)}>
                {label}
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      ))}
    </div>
  )
}
