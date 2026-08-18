import type { TsunaguWorksApplicationStatus } from '@repo/api-client/talent/generated/models'

import { Badge } from '@/components/ui/badge'

// talent 視点のラベル（company 側 #59 と意味が反転する）:
// applied = 相手の返答待ち（secondary）、offered = 自分が動く番（default で目立たせる）
const statusView: Record<
  TsunaguWorksApplicationStatus,
  {
    label: string
    variant: 'default' | 'secondary' | 'outline' | 'destructive'
  }
> = {
  applied: { label: '応募済み・返答待ち', variant: 'secondary' },
  offered: { label: 'オファーが届いています', variant: 'default' },
  accepted: { label: '承諾済み', variant: 'default' },
  rejected: { label: '不採用', variant: 'destructive' },
  withdrawn: { label: '取り下げ済み', variant: 'outline' },
  declined: { label: '辞退済み', variant: 'outline' },
}

export function ApplicationStatusBadge({
  status,
}: {
  status: TsunaguWorksApplicationStatus
}) {
  const view = statusView[status]
  return <Badge variant={view.variant}>{view.label}</Badge>
}
