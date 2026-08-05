'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import { X } from 'lucide-react'
import type { ProjectSearchParams } from '@/lib/projects'

// 適用中の条件を人が読める形に変換する。
// URLのキー（rate_min など）をそのまま見せても意味が伝わらない
function describe(key: string, value: string): string {
  switch (key) {
    case 'skills':
      return `スキル: ${value}`
    case 'rate_min':
      return `時給 ${Number(value).toLocaleString()}円〜`
    case 'rate_max':
      return `時給 〜${Number(value).toLocaleString()}円`
    case 'hours_max':
      return `週${value}時間まで`
    case 'remote_ok':
      return 'フルリモート可'
    case 'q':
      return `キーワード: ${value}`
    default:
      return `${key}: ${value}`
  }
}

// ページ番号は「絞り込み条件」ではないので、チップには出さない
const FILTER_KEYS = [
  'skills',
  'rate_min',
  'rate_max',
  'hours_max',
  'remote_ok',
  'q',
] as const

export function ActiveFilters({ params }: { params: ProjectSearchParams }) {
  const router = useRouter()
  const searchParams = useSearchParams()

  const active = FILTER_KEYS.filter((key) => params[key]).map((key) => ({
    key,
    label: describe(key, params[key] as string),
  }))

  if (active.length === 0) {
    return null
  }

  function remove(key: string) {
    const next = new URLSearchParams(searchParams.toString())
    next.delete(key)
    // 条件を外したら1ページ目に戻す（絞り込みが緩むので件数が変わる）
    next.delete('page')
    router.push(next.toString() ? `?${next.toString()}` : '?')
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-sm text-muted-foreground">適用中:</span>
      {active.map(({ key, label }) => (
        <button
          key={key}
          type="button"
          onClick={() => remove(key)}
          className="flex h-8 items-center gap-1.5 rounded-full border bg-card pl-3 pr-2 text-sm transition-colors hover:border-destructive hover:text-destructive"
        >
          {label}
          {/* アイコンのみでは意味が伝わらないため、何を外すのかを aria-label で補う */}
          <X className="size-3.5" aria-hidden="true" />
          <span className="sr-only">{label}の条件を外す</span>
        </button>
      ))}
      <button
        type="button"
        onClick={() => router.push('?')}
        className="text-sm text-muted-foreground underline hover:text-foreground"
      >
        すべてクリア
      </button>
    </div>
  )
}
