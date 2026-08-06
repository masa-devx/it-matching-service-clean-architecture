import { Badge } from '@/components/ui/badge'
import type { ApplicationStatus } from '@/lib/applications'

// 状態の見せ方を1か所にまとめる（api/application_status.go の遷移表と対応）。
// description は一覧で状態の意味を補足するために使う（「辞退」と「見送り」は
// 文字が似ているうえ、どちらが誰の行為か分かりにくいため）
const statusMeta: Record<
  ApplicationStatus,
  { label: string; description: string; variant: 'default' | 'secondary' | 'outline' | 'destructive' }
> = {
  applied: {
    label: '応募済み',
    description: '企業の返答を待っています',
    variant: 'secondary',
  },
  offered: {
    label: 'オファー中',
    description: '企業から声がかかっています。承諾すると契約成立です',
    variant: 'default',
  },
  accepted: {
    label: '成立',
    description: '双方が合意しました',
    variant: 'default',
  },
  rejected: {
    label: '見送り',
    description: '企業により見送られました',
    variant: 'outline',
  },
  withdrawn: {
    label: '取り下げ',
    description: '自分で取り下げました',
    variant: 'outline',
  },
  declined: {
    label: '辞退',
    description: 'オファーを辞退しました',
    variant: 'outline',
  },
}

export function applicationStatusLabel(status: ApplicationStatus): string {
  return statusMeta[status].label
}

export function applicationStatusDescription(
  status: ApplicationStatus,
): string {
  return statusMeta[status].description
}

export function ApplicationStatusBadge({
  status,
}: {
  status: ApplicationStatus
}) {
  const { label, variant } = statusMeta[status]
  return <Badge variant={variant}>{label}</Badge>
}
