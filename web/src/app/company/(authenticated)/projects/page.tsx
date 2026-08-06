import Link from 'next/link'
import { getMyProjectsWithApplications } from '@/lib/companyProjects'
import { Button } from '@/components/ui/button'
import { CompanyProjectCard } from '@/components/company/CompanyProjectCard'
import { PageHeader } from '@/components/PageHeader'
import { EmptyState } from '@/components/EmptyState'
import { FolderOpen } from 'lucide-react'

export const metadata = { title: '案件管理 | Tsunagu Works' }

export default async function CompanyProjectsPage() {
  const projects = await getMyProjectsWithApplications()

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="案件管理"
        description="掲載した案件と応募状況の一覧です"
        action={
          <Button asChild className="h-11">
            <Link href="/company/projects/new">新規掲載</Link>
          </Button>
        }
      />

      {projects === null ? (
        <p className="text-destructive">案件を取得できませんでした</p>
      ) : projects.length === 0 ? (
        <EmptyState
          icon={FolderOpen}
          title="まだ案件を掲載していません"
          description="案件を掲載すると、人材ユーザーの一覧に表示されます"
          action={
            <Button asChild className="h-11">
              <Link href="/company/projects/new">最初の案件を掲載する</Link>
            </Button>
          }
        />
      ) : (
        <ul className="flex flex-col gap-3">
          {projects.map((project) => (
            <li key={project.id}>
              <CompanyProjectCard project={project} />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
