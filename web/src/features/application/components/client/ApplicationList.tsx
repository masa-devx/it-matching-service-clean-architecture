'use client'

import { useQuery } from '@tanstack/react-query'
import Link from 'next/link'

import { Button } from '@/components/ui/button'

import { ApplicationStatusBadge } from '../ApplicationStatusBadge'
import { ApplicationActions } from './ApplicationActions'
import { myApplicationsQuery } from '../../queries/applications'

export function ApplicationList() {
  const { data: applications, error } = useQuery(myApplicationsQuery)

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
      <div className="space-y-3">
        <p className="text-muted-foreground">まだ応募がありません。</p>
        <Button asChild variant="outline">
          <Link href="/talent/projects">案件を探す</Link>
        </Button>
      </div>
    )
  }

  return (
    <ul className="space-y-3">
      {applications.map((a) => (
        <li key={a.id} className="space-y-2 rounded-lg border p-4">
          <div className="flex items-center justify-between gap-2">
            <Link
              href={`/talent/projects/${a.project_id}`}
              className="font-medium underline-offset-4 hover:underline"
            >
              {a.project_title}
            </Link>
            <ApplicationStatusBadge status={a.status} />
          </div>
          {a.message !== '' && (
            <p className="text-sm whitespace-pre-wrap text-muted-foreground">
              {a.message}
            </p>
          )}
          <p className="text-xs text-muted-foreground">
            応募日: {new Date(a.created_at).toLocaleDateString('ja-JP')}
          </p>
          <ApplicationActions application={a} />
        </li>
      ))}
    </ul>
  )
}
