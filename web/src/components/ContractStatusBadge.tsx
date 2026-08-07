import { Badge } from '@/components/ui/badge'
import type { ContractStatus } from '@/lib/contracts'

// 契約の進行状態。応募の選考状態・案件の掲載状態とは別のドメイン概念なので混ぜない。
// description は「今どういう状態か」を補うために使う——状態名だけでは
// 「検収待ち」が誰の番なのかが伝わらないため
const statusMeta: Record<
  ContractStatus,
  {
    label: string
    description: string
    variant: 'default' | 'secondary' | 'outline'
  }
> = {
  active: {
    label: '成立',
    description: '合意が成立しました。稼働の開始を待っています',
    variant: 'secondary',
  },
  working: {
    label: '稼働中',
    description: '作業が進行中です。週ごとに稼働報告を提出します',
    variant: 'default',
  },
  reviewing: {
    label: '検収待ち',
    description: '企業が成果を確認しています',
    variant: 'default',
  },
  completed: {
    label: '完了',
    description: '検収が完了し、取引が終了しました',
    variant: 'secondary',
  },
  cancelled: {
    label: '中止',
    description: '取引は中止されました',
    variant: 'outline',
  },
}

export function contractStatusLabel(status: ContractStatus): string {
  return statusMeta[status].label
}

export function contractStatusDescription(status: ContractStatus): string {
  return statusMeta[status].description
}

export function ContractStatusBadge({ status }: { status: ContractStatus }) {
  const { label, variant } = statusMeta[status]
  return <Badge variant={variant}>{label}</Badge>
}
