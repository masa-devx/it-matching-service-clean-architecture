import Link from 'next/link'
import { PROJECT_STATUSES, type ProjectStatus } from '@/lib/companyProjects'
import { projectStatusLabel } from '@/components/company/ProjectStatusBadge'

// 掲載状態での絞り込み。リンクなので RSC のまま完結し、条件が URL に残る
export function ProjectStatusFilter({
  current,
}: {
  current: string | undefined
}) {
  return (
    <nav aria-label="掲載状態で絞り込み" className="flex flex-wrap gap-2">
      <FilterLink current={current} value="" label="すべて" />
      {PROJECT_STATUSES.map((status) => (
        <FilterLink
          key={status}
          current={current}
          value={status}
          label={projectStatusLabel(status)}
        />
      ))}
    </nav>
  )
}

// 選択中は aria-current で伝える（色だけに頼らない）
function FilterLink({
  current,
  value,
  label,
}: {
  current: string | undefined
  value: ProjectStatus | ''
  label: string
}) {
  const active = (current ?? '') === value

  return (
    <Link
      href={value ? `/company/projects?status=${value}` : '/company/projects'}
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
