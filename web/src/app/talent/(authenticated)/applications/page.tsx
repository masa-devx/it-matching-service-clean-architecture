import Link from 'next/link'
import { FileSearch } from 'lucide-react'
import {
  getMyApplications,
  currentApplicationPage,
  APPLICATIONS_PER_PAGE,
  type ApplicationSearchParams,
} from '@/lib/applications'
import {
  ApplicationStatusBadge,
  applicationStatusDescription,
} from '@/components/ApplicationStatusBadge'
import { ApplicationStatusFilter } from '@/components/ApplicationStatusFilter'
import { ApplicationActions } from '@/components/talent/ApplicationActions'
import { PageHeader } from '@/components/PageHeader'
import { EmptyState } from '@/components/EmptyState'
import { Pagination } from '@/components/Pagination'
import { Button } from '@/components/ui/button'

export const metadata = { title: '応募履歴 | Tsunagu Works' }

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

// searchParams は Promise（Next.js 15+ の動的APIの非同期化）
export default async function TalentApplicationsPage({
  searchParams,
}: {
  searchParams: Promise<ApplicationSearchParams>
}) {
  const params = await searchParams
  const result = await getMyApplications(params)

  if (!result) {
    return <p className="text-destructive">応募履歴を取得できませんでした</p>
  }

  const page = currentApplicationPage(params)

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="応募履歴"
        description={`${result.total.toLocaleString()}件の応募`}
      />

      <ApplicationStatusFilter
        basePath="/talent/applications"
        current={params.status}
      />

      {result.applications.length === 0 ? (
        <EmptyState
          icon={FileSearch}
          title="応募がありません"
          description="気になる案件に応募すると、ここで選考状況を確認できます"
          action={
            <Button asChild className="h-11">
              <Link href="/talent/projects">案件を探す</Link>
            </Button>
          }
        />
      ) : (
        <ul className="flex flex-col gap-3">
          {result.applications.map((application) => (
            <li
              key={application.id}
              className="flex flex-col gap-4 rounded-lg border bg-card p-6"
            >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="flex flex-col gap-1">
                  <Link
                    href={`/talent/projects/${application.project_id}`}
                    className="font-bold hover:underline"
                  >
                    {application.project_title}
                  </Link>
                  <p className="text-sm text-muted-foreground">
                    {application.company_name}・
                    {formatDate(application.created_at)}に応募
                  </p>
                </div>
                <ApplicationStatusBadge status={application.status} />
              </div>

              <p className="text-sm text-muted-foreground">
                {applicationStatusDescription(application.status)}
              </p>

              {/* 志望動機は長くなるため折りたたむ（一覧の一覧性を優先） */}
              <details className="text-sm">
                <summary className="w-fit cursor-pointer text-muted-foreground hover:text-foreground">
                  送信した志望動機
                </summary>
                <p className="mt-2 whitespace-pre-wrap break-words leading-relaxed">
                  {application.message}
                </p>
              </details>

              <ApplicationActions
                applicationId={application.id}
                status={application.status}
              />
            </li>
          ))}
        </ul>
      )}

      <Pagination
        currentPage={page}
        total={result.total}
        perPage={APPLICATIONS_PER_PAGE}
      />
    </div>
  )
}
