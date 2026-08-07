import Link from 'next/link'
import { FolderOpen } from 'lucide-react'
import {
  getMyProjects,
  currentMyProjectPage,
  MY_PROJECTS_PER_PAGE,
  type MyProjectSearchParams,
} from '@/lib/companyProjects'
import { Button } from '@/components/ui/button'
import { CompanyProjectCard } from '@/components/company/CompanyProjectCard'
import { ProjectStatusFilter } from '@/components/company/ProjectStatusFilter'
import { PageHeader } from '@/components/PageHeader'
import { EmptyState } from '@/components/EmptyState'
import { Pagination } from '@/components/Pagination'

export const metadata = { title: '案件管理 | Tsunagu Works' }

export default async function CompanyProjectsPage({
  searchParams,
}: {
  searchParams: Promise<MyProjectSearchParams>
}) {
  const params = await searchParams
  const result = await getMyProjects(params)

  if (!result) {
    return <p className="text-destructive">案件を取得できませんでした</p>
  }

  const page = currentMyProjectPage(params)
  // 絞り込み中に0件なら「まだ掲載していない」ではなく「条件に合わない」
  const filtered = Boolean(params.status)

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="案件管理"
        description={`${result.total.toLocaleString()}件の案件（下書き・募集終了を含む）`}
        action={
          <Button asChild className="h-11">
            <Link href="/company/projects/new">新規掲載</Link>
          </Button>
        }
      />

      <ProjectStatusFilter current={params.status} />

      {result.projects.length === 0 ? (
        filtered ? (
          <EmptyState
            icon={FolderOpen}
            title="この状態の案件はありません"
            description="別の状態を選ぶか、すべて表示してください"
            action={
              <Button asChild variant="outline" className="h-11">
                <Link href="/company/projects">すべて表示</Link>
              </Button>
            }
          />
        ) : (
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
        )
      ) : (
        <ul className="flex flex-col gap-3">
          {result.projects.map((project) => (
            <li key={project.id}>
              <CompanyProjectCard project={project} />
            </li>
          ))}
        </ul>
      )}

      <Pagination
        currentPage={page}
        total={result.total}
        perPage={MY_PROJECTS_PER_PAGE}
      />
    </div>
  )
}
