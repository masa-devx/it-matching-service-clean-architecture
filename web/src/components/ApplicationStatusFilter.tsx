import Link from 'next/link'
import {
  APPLICATION_STATUSES,
  type ApplicationStatus,
} from '@/lib/applications'
import { applicationStatusLabel } from '@/components/ApplicationStatusBadge'

// 選考状態での絞り込み。リンクで実装することで RSC のまま完結し、
// 条件が URL に残る（リロード・共有・ブラウザバックで再現できる）。
// 人材の応募履歴と企業の応募者管理で共用する（basePath だけ差し替える）
export function ApplicationStatusFilter({
  basePath,
  current,
}: {
  basePath: string
  current: string | undefined
}) {
  return (
    <nav aria-label="選考状態で絞り込み" className="flex flex-wrap gap-2">
      <FilterLink
        basePath={basePath}
        current={current}
        value=""
        label="すべて"
      />
      {APPLICATION_STATUSES.map((status) => (
        <FilterLink
          key={status}
          basePath={basePath}
          current={current}
          value={status}
          label={applicationStatusLabel(status)}
        />
      ))}
    </nav>
  )
}

// 選択中は aria-current で伝える（色だけに頼らない）
function FilterLink({
  basePath,
  current,
  value,
  label,
}: {
  basePath: string
  current: string | undefined
  value: ApplicationStatus | ''
  label: string
}) {
  const active = (current ?? '') === value

  return (
    <Link
      href={value ? `${basePath}?status=${value}` : basePath}
      aria-current={active ? 'page' : undefined}
      className={`rounded-full border px-4 py-2 text-sm transition-colors ${
        active
          ? 'border-primary bg-primary text-primary-foreground'
          : 'hover:bg-muted'
      }`}
    >
      {label}
    </Link>
  )
}
