import { Badge } from '@/components/ui/badge'
import type { WorkReportStatus } from '@/lib/workReports'

// 稼働報告の状態。契約の状態とは別物なので混ぜない
// （契約の「検収待ち」は仕事全体の確認、報告の「確認待ち」はその週の作業の確認）
const statusMeta: Record<
  WorkReportStatus,
  {
    label: string
    description: string
    variant: 'default' | 'secondary' | 'outline'
  }
> = {
  submitted: {
    label: '確認待ち',
    description: '企業の確認を待っています',
    variant: 'default',
  },
  approved: {
    label: '承認済み',
    description: '企業が内容を確認しました',
    variant: 'secondary',
  },
  rejected: {
    label: '差し戻し',
    description: '修正して再提出してください',
    variant: 'outline',
  },
}

export function workReportStatusLabel(status: WorkReportStatus): string {
  return statusMeta[status].label
}

export function workReportStatusDescription(status: WorkReportStatus): string {
  return statusMeta[status].description
}

export function WorkReportStatusBadge({
  status,
}: {
  status: WorkReportStatus
}) {
  const { label, variant } = statusMeta[status]
  return <Badge variant={variant}>{label}</Badge>
}
