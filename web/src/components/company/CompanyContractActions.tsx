'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { updateContractStatus } from '@/lib/contractClient'
import type { ContractStatus } from '@/lib/contracts'
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
  status: ContractStatus
  label: string
  doneLabel: string
  variant?: 'default' | 'outline'
  confirm: { title: string; description: string }
}

// 状態ごとに「企業が実行できる操作」を並べる。
// api/contract_status.go の遷移表のうち roleCompany の行と対応させること。
//
// 人材側（ContractActions）と足すと遷移表の全行になる:
//   人材 = 稼働開始・検収依頼・中止 / 企業 = 検収・差し戻し・中止
// 稼働の開始と検収の依頼が企業側に無いのは、どちらも「作業する人の申告」だから
const actionsByStatus: Partial<Record<ContractStatus, Action[]>> = {
  active: [
    {
      status: 'cancelled',
      label: '契約を中止する',
      doneLabel: '契約を中止し',
      variant: 'outline',
      confirm: {
        title: 'この契約を中止しますか？',
        description:
          '人材にも中止として伝わります。取り消して再開することはできません。',
      },
    },
  ],
  working: [
    {
      status: 'cancelled',
      label: '契約を中止する',
      doneLabel: '契約を中止し',
      variant: 'outline',
      confirm: {
        title: 'この契約を中止しますか？',
        description:
          '人材にも中止として伝わります。提出済みの稼働報告は記録として残りますが、取り消して再開することはできません。',
      },
    },
  ],
  reviewing: [
    {
      status: 'completed',
      label: '検収して完了する',
      doneLabel: '検収を完了し',
      confirm: {
        title: 'この契約を完了しますか？',
        description:
          '成果を確認したものとして取引を終了します。完了した契約は元に戻せません。内容に不備があれば、完了ではなく差し戻してください。',
      },
    },
    {
      status: 'working',
      label: '稼働中に差し戻す',
      doneLabel: '稼働中に差し戻し',
      variant: 'outline',
      confirm: {
        title: '稼働中に差し戻しますか？',
        description:
          '作業を続けてもらうために契約を稼働中に戻します。人材は稼働報告を再び提出できるようになります。',
      },
    },
    {
      status: 'cancelled',
      label: '契約を中止する',
      doneLabel: '契約を中止し',
      variant: 'outline',
      confirm: {
        title: 'この契約を中止しますか？',
        description:
          '検収を行わずに取引を終了します。取り消して再開することはできません。',
      },
    },
  ],
}

export function CompanyContractActions({
  contractId,
  status,
}: {
  contractId: number
  status: ContractStatus
}) {
  const router = useRouter()
  const [pending, startTransition] = useTransition()
  const [sending, setSending] = useState(false)

  const actions = actionsByStatus[status]
  if (!actions) {
    return null
  }

  async function run(next: ContractStatus, doneLabel: string) {
    setSending(true)
    const result = await updateContractStatus(contractId, next)
    setSending(false)

    if (!result.ok) {
      toast.error(result.error)
      // 人材が先に操作していた場合（409）は、画面を最新に合わせる
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
