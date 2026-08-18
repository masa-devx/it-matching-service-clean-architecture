import type { TsunaguWorksApplicationStatus } from '@repo/api-client/company/generated/models'

import { Badge } from '@/components/ui/badge'

// 6状態の表示定義（Record で網羅を型強制・状態が増えたらコンパイルエラーで気づく）。
// company 視点のラベル: 「選考待ち」= 自分が動く番、「オファー中」= 相手の返答待ち
const statusView: Record<
  TsunaguWorksApplicationStatus,
  {
    label: string
    variant: 'default' | 'secondary' | 'outline' | 'destructive'
  }
> = {
  applied: { label: '選考待ち', variant: 'default' },
  offered: { label: 'オファー中', variant: 'secondary' },
  accepted: { label: '承諾済み', variant: 'default' },
  rejected: { label: '不採用', variant: 'destructive' },
  withdrawn: { label: '取り下げ', variant: 'outline' },
  declined: { label: '辞退', variant: 'outline' },
}

export function ApplicationStatusBadge({
  status,
}: {
  status: TsunaguWorksApplicationStatus
}) {
  const view = statusView[status]
  return <Badge variant={view.variant}>{view.label}</Badge>
}
