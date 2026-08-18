'use client'

import type { TsunaguWorksProject } from '@repo/api-client/company/generated/models'
import { useQuery } from '@tanstack/react-query'
import {
  CalendarDays,
  Clock,
  FolderPlus,
  JapaneseYen,
  Wifi,
  WifiOff,
} from 'lucide-react'
import Link from 'next/link'

import { EmptyState } from '@/components/EmptyState'
import { SkillBadges } from '@/components/SkillBadges'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

import { ProjectStatusBadge } from '../ProjectStatusBadge'
import { ProjectStatusActions } from './ProjectStatusActions'
import { companyProjectsQuery } from '../../queries/companyProjects'

function rateText(p: TsunaguWorksProject): string {
  if (p.hourly_rate_min == null && p.hourly_rate_max == null) {
    return '時給未設定'
  }
  const min = p.hourly_rate_min != null ? `${p.hourly_rate_min}円` : ''
  const max = p.hourly_rate_max != null ? `${p.hourly_rate_max}円` : ''
  return `${min}〜${max}`
}

export function ProjectList() {
  const { data: projects, error } = useQuery(companyProjectsQuery)

  if (error) {
    return (
      <p role="alert" className="text-destructive">
        {error.message}
      </p>
    )
  }
  if (!projects) {
    return null
  }
  if (projects.length === 0) {
    return (
      <EmptyState
        icon={FolderPlus}
        title="案件がまだありません"
        description="最初の案件を掲載して、人材からの応募を受け付けましょう。"
      >
        <Button asChild size="sm">
          <Link href="/company/projects/new">新規作成</Link>
        </Button>
      </EmptyState>
    )
  }

  return (
    <ul className="space-y-3">
      {projects.map((p) => (
        <li key={p.id}>
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between gap-2">
                <CardTitle>{p.title}</CardTitle>
                <ProjectStatusBadge status={p.status} />
              </div>
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
            <CardFooter className="flex flex-wrap items-center justify-between gap-2">
              <div className="flex gap-2">
                <Button asChild variant="outline" size="sm">
                  <Link href={`/company/projects/${p.id}/applications`}>
                    応募を見る
                  </Link>
                </Button>
                <Button asChild variant="outline" size="sm">
                  <Link href={`/company/projects/${p.id}/edit`}>編集</Link>
                </Button>
              </div>
              <ProjectStatusActions project={p} />
            </CardFooter>
          </Card>
        </li>
      ))}
    </ul>
  )
}
