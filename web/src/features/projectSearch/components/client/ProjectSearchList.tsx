'use client'

import type { TsunaguWorksProject } from '@repo/api-client/talent/generated/models'
import { useInfiniteQuery } from '@tanstack/react-query'
import Link from 'next/link'

import { Button } from '@/components/ui/button'

import {
  projectSearchQuery,
  type ProjectSearchFilters,
} from '../../queries/projects'

// features/project（company 視点）にも同種の関数があるが import しない（feature 間依存禁止）。
// 表示の要件は視点ごとに別々に育つため、小さな複製を許容する
function rateText(p: TsunaguWorksProject): string {
  if (p.hourly_rate_min == null && p.hourly_rate_max == null) {
    return '時給未設定'
  }
  const min = p.hourly_rate_min != null ? `${p.hourly_rate_min}円` : ''
  const max = p.hourly_rate_max != null ? `${p.hourly_rate_max}円` : ''
  return `時給 ${min}〜${max}`
}

export function ProjectSearchList({
  filters,
}: {
  filters: ProjectSearchFilters
}) {
  const { data, error, hasNextPage, fetchNextPage, isFetchingNextPage } =
    useInfiniteQuery(projectSearchQuery(filters))

  if (error) {
    return (
      <p role="alert" className="text-destructive">
        {error.message}
      </p>
    )
  }
  if (!data) {
    return null
  }

  // pages = [1ページ目, 2ページ目, ...] を1本のリストに畳む
  const projects = data.pages.flatMap((page) => page.projects)

  if (projects.length === 0) {
    return (
      <p className="text-muted-foreground">
        条件に合う案件が見つかりませんでした。条件を減らして試してください。
      </p>
    )
  }

  return (
    <div className="space-y-4">
      <ul className="space-y-3">
        {projects.map((p) => (
          <li key={p.id} className="rounded-lg border p-4">
            <Link
              href={`/talent/projects/${p.id}`}
              className="font-medium underline-offset-4 hover:underline"
            >
              {p.title}
            </Link>
            <p className="mt-1 text-sm text-muted-foreground">
              {rateText(p)} ・ 週{p.hours_per_week}時間 ・{' '}
              {p.remote_ok ? 'リモート可' : 'リモート不可'} ・{' '}
              {new Date(p.created_at).toLocaleDateString('ja-JP')}
            </p>
            {p.required_skills.length > 0 && (
              <p className="mt-1 text-sm text-muted-foreground">
                スキル: {p.required_skills.join(' / ')}
              </p>
            )}
          </li>
        ))}
      </ul>

      {hasNextPage && (
        <div className="flex justify-center">
          <Button
            variant="outline"
            onClick={() => fetchNextPage()}
            disabled={isFetchingNextPage}
          >
            {isFetchingNextPage ? '読み込み中…' : 'もっと見る'}
          </Button>
        </div>
      )}
    </div>
  )
}
