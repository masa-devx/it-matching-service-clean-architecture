'use client'

import { useQuery } from '@tanstack/react-query'

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
      <p className="text-muted-foreground">
        まだ応募がありません。案件が公開中であれば、人材の検索結果に表示されています。
      </p>
    )
  }

  return (
    <ul className="space-y-3">
      {applications.map((a) => (
        <li key={a.id} className="space-y-2 rounded-lg border p-4">
          <div className="flex items-center justify-between gap-2">
            <span className="font-medium">{a.talent_display_name}</span>
            <ApplicationStatusBadge status={a.status} />
          </div>
          {a.talent_skills.length > 0 && (
            <p className="text-sm text-muted-foreground">
              スキル: {a.talent_skills.join(' / ')}
            </p>
          )}
          {a.message !== '' && (
            <p className="text-sm whitespace-pre-wrap">{a.message}</p>
          )}
          <p className="text-xs text-muted-foreground">
            応募日: {new Date(a.created_at).toLocaleDateString('ja-JP')}
          </p>
          <SelectionActions application={a} />
        </li>
      ))}
    </ul>
  )
}
