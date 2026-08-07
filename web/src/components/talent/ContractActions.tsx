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
  // 相手に影響する操作だけ確認を挟む（デザインシステムの規約に従う）
  confirm?: { title: string; description: string }
}

// 状態ごとに「人材が実行できる操作」を並べる。
// api/contract_status.go の遷移表のうち roleTalent の行と対応させること
// （ここに無い操作は画面から実行できないが、正しさの担保は API 側の 409）。
//
// 終端（completed / cancelled）はキーを持たない＝操作が出ない
const actionsByStatus: Partial<Record<ContractStatus, Action[]>> = {
  active: [
    {
      status: 'working',
      label: '稼働を開始する',
      doneLabel: '稼働を開始し',
      // 開始は相手に不利益が無く、実質的に取り消せる（中止できる）ため確認しない
    },
    {
      status: 'cancelled',
      label: '契約を中止する',
      doneLabel: '契約を中止し',
      variant: 'outline',
      confirm: {
        title: 'この契約を中止しますか？',
        description:
          '企業にも中止として伝わります。取り消して再開することはできません。',
      },
    },
  ],
  working: [
    {
      status: 'reviewing',
      label: '検収を依頼する',
      doneLabel: '検収を依頼し',
      confirm: {
        title: '検収を依頼しますか？',
        description:
          '作業が完了したものとして企業に確認を求めます。確認中は稼働報告を追加できません（不備があれば企業から差し戻されます）。',
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
          '企業にも中止として伝わります。提出済みの稼働報告は記録として残りますが、取り消して再開することはできません。',
      },
    },
  ],
  reviewing: [
    {
      status: 'cancelled',
      label: '契約を中止する',
      doneLabel: '契約を中止し',
      variant: 'outline',
      confirm: {
        title: 'この契約を中止しますか？',
        description:
          '企業が検収中です。中止すると取引は完了せず、取り消して再開することはできません。',
      },
    },
  ],
}

export function ContractActions({
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
      // 相手が先に操作していた場合（409）は、画面を最新に合わせる
      startTransition(() => router.refresh())
      return
    }

    toast.success(`${doneLabel}ました`)
    startTransition(() => router.refresh())
  }

  const disabled = sending || pending

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
