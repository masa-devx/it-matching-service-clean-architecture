import type { TsunaguWorksProjectStatus } from '@repo/api-client/company/generated/models'

import { Badge } from '@/components/ui/badge'

// 状態の表示定義はこの1か所（ラベルと見た目のペアで管理する）。
// props だけで動く表示部品なので Server / Client どちらからも使える
const statusView: Record<
  TsunaguWorksProjectStatus,
  { label: string; variant: 'default' | 'secondary' | 'outline' }
> = {
  draft: { label: '下書き', variant: 'secondary' },
  published: { label: '公開中', variant: 'default' },
  closed: { label: '募集終了', variant: 'outline' },
}

export function ProjectStatusBadge({
  status,
}: {
  status: TsunaguWorksProjectStatus
}) {
  const view = statusView[status]
  return <Badge variant={view.variant}>{view.label}</Badge>
}
