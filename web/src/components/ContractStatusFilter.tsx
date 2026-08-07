import Link from 'next/link'
import { CONTRACT_STATUSES, type ContractStatus } from '@/lib/contracts'
import { contractStatusLabel } from '@/components/ContractStatusBadge'

// 契約状態での絞り込み。リンクなので RSC のまま完結し、条件が URL に残る。
// 人材（/talent/contracts）と企業（/company/contracts）で共用する
export function ContractStatusFilter({
  basePath,
  current,
}: {
  basePath: string
  current: string | undefined
}) {
  return (
    <nav aria-label="契約状態で絞り込み" className="flex flex-wrap gap-2">
      <FilterLink
        basePath={basePath}
        current={current}
        value=""
        label="すべて"
      />
      {CONTRACT_STATUSES.map((status) => (
        <FilterLink
          key={status}
          basePath={basePath}
          current={current}
          value={status}
          label={contractStatusLabel(status)}
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
  value: ContractStatus | ''
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
