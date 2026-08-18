'use client'

import type { TsunaguWorksProject } from '@repo/api-client/talent/generated/models'
import { useInfiniteQuery } from '@tanstack/react-query'
import { CalendarDays, Clock, JapaneseYen, Wifi, WifiOff } from 'lucide-react'
import Link from 'next/link'

import { SkillBadges } from '@/components/SkillBadges'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

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
  return `${min}〜${max}`
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
      {/* 探索系はグリッド（md で2カラム）: 一覧性を優先。操作を持つ company 側は1カラムのまま */}
      <ul className="grid gap-3 md:grid-cols-2">
        {projects.map((p) => (
          <li key={p.id}>
            <Card className="h-full transition-shadow hover:shadow-md">
              <CardHeader>
                <CardTitle>
                  <Link
                    href={`/talent/projects/${p.id}`}
                    className="underline-offset-4 hover:underline"
                  >
                    {p.title}
                  </Link>
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm text-muted-foreground">
                  <span className="flex items-center gap-1">
                    <JapaneseYen className="size-3.5" aria-hidden="true" />
                    {rateText(p)}
                  </span>
                  <span className="flex items-center gap-1">
                    <Clock className="size-3.5" aria-hidden="true" />週
                    {p.hours_per_week}時間
                  </span>
                  <span className="flex items-center gap-1">
                    {p.remote_ok ? (
                      <Wifi className="size-3.5" aria-hidden="true" />
                    ) : (
                      <WifiOff className="size-3.5" aria-hidden="true" />
                    )}
                    {p.remote_ok ? 'リモート可' : 'リモート不可'}
                  </span>
                  <span className="flex items-center gap-1">
                    <CalendarDays className="size-3.5" aria-hidden="true" />
                    {new Date(p.created_at).toLocaleDateString('ja-JP')}
                  </span>
                </div>
                <SkillBadges skills={p.required_skills} />
              </CardContent>
            </Card>
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
