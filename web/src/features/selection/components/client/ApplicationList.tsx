'use client'

import { useQuery } from '@tanstack/react-query'
import { CalendarDays, Inbox } from 'lucide-react'

import { EmptyState } from '@/components/EmptyState'
import { ExpandableText } from '@/components/ExpandableText'
import { SkillBadges } from '@/components/SkillBadges'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

import { ApplicationStatusBadge } from '../ApplicationStatusBadge'
import { SelectionActions } from './SelectionActions'
import { applicationsQuery } from '../../queries/applications'

export function ApplicationList({ projectId }: { projectId: number }) {
  const { data: applications, error } = useQuery(applicationsQuery(projectId))

  if (error) {
    return (
      <p role="alert" className="text-destructive">
        {error.message}
      </p>
    )
  }
  if (!applications) {
    return null
  }
  if (applications.length === 0) {
    return (
      <EmptyState
        icon={Inbox}
        title="まだ応募がありません"
        description="案件が公開中であれば、人材の検索結果に表示されています。"
      />
    )
  }

  return (
    <ul className="space-y-3">
      {applications.map((a) => (
        <li key={a.id}>
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between gap-2">
                <CardTitle>{a.talent_display_name}</CardTitle>
                <ApplicationStatusBadge status={a.status} />
              </div>
            </CardHeader>
            <CardContent className="space-y-3">
              <SkillBadges skills={a.talent_skills} />
              {a.message !== '' && <ExpandableText text={a.message} />}
              <p className="flex items-center gap-1 text-xs text-muted-foreground">
                <CalendarDays className="size-3.5" aria-hidden="true" />
                応募日: {new Date(a.created_at).toLocaleDateString('ja-JP')}
              </p>
              <SelectionActions application={a} />
            </CardContent>
          </Card>
        </li>
      ))}
    </ul>
  )
}
