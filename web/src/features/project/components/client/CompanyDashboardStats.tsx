'use client'

import { useQuery } from '@tanstack/react-query'

import { StatCard } from '@/components/StatCard'

import { companyProjectsQuery } from '../../queries/companyProjects'

// 件数は一覧キャッシュからの導出（derived state）: filter().length を描画時に計算するだけで、
// 別 state に持たない＝invalidate（公開・終了の操作後）に自動で追随する。
// 専用の集計 API は作らない（一覧が持てなくなってから・#60 と同じ判断）
export function CompanyDashboardStats() {
  const { data: projects } = useQuery(companyProjectsQuery)
  if (!projects) {
    return null
  }

  const countBy = (status: string) =>
    projects.filter((p) => p.status === status).length

  return (
    <div className="grid gap-3 sm:grid-cols-3">
      <StatCard
        label="掲載中の案件"
        value={countBy('published')}
        href="/company/projects"
      />
      <StatCard
        label="下書き"
        value={countBy('draft')}
        href="/company/projects"
      />
      <StatCard
        label="募集終了"
        value={countBy('closed')}
        href="/company/projects"
      />
    </div>
  )
}
