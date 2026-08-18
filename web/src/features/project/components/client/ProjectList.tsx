'use client'

import { useQuery } from '@tanstack/react-query'
import type { TsunaguWorksProject } from '@repo/api-client/company/generated/models'
import Link from 'next/link'

import { Button } from '@/components/ui/button'

import { ProjectStatusBadge } from '../ProjectStatusBadge'
import { companyProjectsQuery } from '../../queries/companyProjects'

function rateText(p: TsunaguWorksProject): string {
  if (p.hourly_rate_min == null && p.hourly_rate_max == null) {
    return '時給未設定'
  }
  const min = p.hourly_rate_min != null ? `${p.hourly_rate_min}円` : ''
  const max = p.hourly_rate_max != null ? `${p.hourly_rate_max}円` : ''
  return `時給 ${min}〜${max}`
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
      <p className="text-muted-foreground">
        案件がまだありません。「新規作成」から最初の案件を掲載しましょう。
      </p>
    )
  }

  return (
    <ul className="space-y-3">
      {projects.map((p) => (
        <li key={p.id} className="rounded-lg border p-4">
          <div className="flex items-center justify-between gap-2">
            <span className="font-medium">{p.title}</span>
            <div className="flex items-center gap-2">
              <ProjectStatusBadge status={p.status} />
              <Button asChild variant="outline" size="sm">
                <Link href={`/company/projects/${p.id}/edit`}>編集</Link>
              </Button>
            </div>
          </div>
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
  )
}
