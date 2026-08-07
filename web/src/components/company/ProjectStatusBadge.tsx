import { Badge } from '@/components/ui/badge'
import type { ProjectStatus } from '@/lib/companyProjects'

// 案件の掲載状態。応募の選考状態（ApplicationStatusBadge）とは別のドメイン概念なので混ぜない。
// Record にすることで、状態を足したときに定義漏れがコンパイルエラーになる
const statusMeta: Record<
  ProjectStatus,
  {
    label: string
    description: string
    variant: 'default' | 'secondary' | 'outline'
  }
> = {
  draft: {
    label: '下書き',
    description: '人材には表示されません',
    variant: 'outline',
  },
  published: {
    label: '公開中',
    description: '人材の一覧に表示され、応募を受け付けています',
    variant: 'default',
  },
  closed: {
    label: '募集終了',
    description: '新しい応募は受け付けていません',
    variant: 'secondary',
  },
}

export function projectStatusLabel(status: ProjectStatus): string {
  return statusMeta[status].label
}

export function projectStatusDescription(status: ProjectStatus): string {
  return statusMeta[status].description
}

export function ProjectStatusBadge({ status }: { status: ProjectStatus }) {
  const { label, variant } = statusMeta[status]
  return <Badge variant={variant}>{label}</Badge>
}
