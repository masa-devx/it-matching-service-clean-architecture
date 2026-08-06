import Link from 'next/link'
import { notFound } from 'next/navigation'
import { ArrowLeft, Users } from 'lucide-react'
import { getProjectApplications } from '@/lib/companyApplications'
import {
  currentApplicationPage,
  APPLICATIONS_PER_PAGE,
  type ApplicationSearchParams,
} from '@/lib/applications'
import { getProject } from '@/lib/projects'
import { ApplicantCard } from '@/components/company/ApplicantCard'
import { ApplicationStatusFilter } from '@/components/ApplicationStatusFilter'
import { PageHeader } from '@/components/PageHeader'
import { EmptyState } from '@/components/EmptyState'
import { Pagination } from '@/components/Pagination'

export const metadata = { title: '応募者管理 | Tsunagu Works' }

export default async function CompanyApplicantsPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>
  searchParams: Promise<ApplicationSearchParams>
}) {
  const [{ id }, query] = await Promise.all([params, searchParams])
  const projectId = Number(id)

  const result = await getProjectApplications(projectId, query)
  // 他社の案件・存在しない案件は API が404を返す（存在を漏らさない方針）
  if (!result) {
    notFound()
  }

  // 案件名は見出しの補足。公開中の案件しか引けないため、取れなくても一覧は表示する
  // （企業向けの案件詳細APIが無いことによる制約）
  const project = await getProject(projectId)
  const page = currentApplicationPage(query)

  return (
    <div className="flex flex-col gap-6">
      <Link
        href="/company/projects"
        className="flex w-fit items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="size-4" aria-hidden="true" />
        案件管理に戻る
      </Link>

      <PageHeader
        title="応募者管理"
        description={
          project
            ? `${project.title}・${result.total.toLocaleString()}件の応募`
            : `${result.total.toLocaleString()}件の応募`
        }
      />

      <ApplicationStatusFilter
        basePath={`/company/projects/${projectId}/applications`}
        current={query.status}
      />

      {result.applications.length === 0 ? (
        <EmptyState
          icon={Users}
          title="応募がありません"
          description="案件が公開されていれば、人材からの応募がここに表示されます"
        />
      ) : (
        <ul className="flex flex-col gap-3">
          {result.applications.map((application) => (
            <li key={application.id}>
              <ApplicantCard application={application} />
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
