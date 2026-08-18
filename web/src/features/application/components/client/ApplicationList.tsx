'use client'

import { useQuery } from '@tanstack/react-query'
import { CalendarDays } from 'lucide-react'
import Link from 'next/link'

import { ExpandableText } from '@/components/ExpandableText'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

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
        <li key={a.id}>
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between gap-2">
                <CardTitle>
                  <Link
                    href={`/talent/projects/${a.project_id}`}
                    className="underline-offset-4 hover:underline"
                  >
                    {a.project_title}
                  </Link>
                </CardTitle>
                <ApplicationStatusBadge status={a.status} />
              </div>
            </CardHeader>
            <CardContent className="space-y-3">
              {a.message !== '' && <ExpandableText text={a.message} />}
              <p className="flex items-center gap-1 text-xs text-muted-foreground">
                <CalendarDays className="size-3.5" aria-hidden="true" />
                応募日: {new Date(a.created_at).toLocaleDateString('ja-JP')}
              </p>
              {/* 終端状態では null（フッターを使わないのは空の余白を作らないため） */}
              <ApplicationActions application={a} />
            </CardContent>
          </Card>
        </li>
      ))}
    </ul>
  )
}
