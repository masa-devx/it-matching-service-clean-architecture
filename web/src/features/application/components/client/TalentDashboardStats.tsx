'use client'

import { useQuery } from '@tanstack/react-query'

import { StatCard } from '@/components/StatCard'

import { myApplicationsQuery } from '../../queries/applications'

// 件数は応募一覧キャッシュからの導出（company 側と同じ derived state の型）。
// オファーあり > 0 は「今やるべきこと」なので強調する
export function TalentDashboardStats() {
  const { data: applications } = useQuery(myApplicationsQuery)
  if (!applications) {
    return null
  }

  const applied = applications.filter((a) => a.status === 'applied').length
  const offered = applications.filter((a) => a.status === 'offered').length

  return (
    <div className="grid gap-3 sm:grid-cols-2">
      <StatCard
        label="オファーが届いています"
        value={offered}
        href="/talent/applications"
        highlight={offered > 0}
      />
      <StatCard
        label="応募中（返答待ち）"
        value={applied}
        href="/talent/applications"
      />
    </div>
  )
}
