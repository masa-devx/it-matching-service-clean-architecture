'use client'

import type {
  TsunaguWorksApplicationForCompany,
  TsunaguWorksApplicationStatus,
} from '@repo/api-client/company/generated/models'
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
import type { SelectionAction } from '@/external/handler/selection'

import { changeSelectionAction } from '../../actions/status'
import { selectionKeys } from '../../queries/applications'

// 状態ごとに出す操作（actor=company の遷移表の UI 側の写し）。
// company が動けるのは applied のときだけ。正しさの保証はサーバー（条件付きUPDATE）で、
// タブ2枚などで食い違ったら 409（現在: x）がトーストで返る。
// どちらも相手に届く不可逆の意思表示なので、両方とも確認ダイアログを挟む
const actionsByStatus: Record<
  TsunaguWorksApplicationStatus,
  Array<{
    action: SelectionAction
    label: string
    destructive?: boolean
    confirm: { title: string; description: string }
  }>
> = {
  applied: [
    {
      action: 'offer',
      label: 'オファーする',
      confirm: {
        title: 'この応募者にオファーしますか？',
        description: 'オファーすると相手が承諾・辞退を選べるようになります。',
      },
    },
    {
      action: 'reject',
      label: '不採用にする',
      destructive: true,
      confirm: {
        title: 'この応募を不採用にしますか？',
        description: '不採用にすると取り消せません。',
      },
    },
  ],
  offered: [],
  accepted: [],
  rejected: [],
  withdrawn: [],
  declined: [],
}

export function SelectionActions({
  application,
}: {
  application: TsunaguWorksApplicationForCompany
}) {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: async (action: SelectionAction) => {
      const result = await changeSelectionAction(application.id, action)
      if (!result.ok) {
        throw new Error(result.error)
      }
      return result.data
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: selectionKeys.all })
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
      {actions.map(({ action, label, destructive, confirm }) => (
        <AlertDialog key={action}>
          <AlertDialogTrigger asChild>
            <Button
              variant={destructive ? 'destructive' : 'outline'}
              size="sm"
              disabled={mutation.isPending}
            >
              {label}
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{confirm.title}</AlertDialogTitle>
              <AlertDialogDescription>
                {application.talent_display_name} さん — {confirm.description}
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
