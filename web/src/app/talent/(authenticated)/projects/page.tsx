import {
  searchProjects,
  currentPage,
  PROJECTS_PER_PAGE,
  type ProjectSearchParams,
} from '@/lib/projects'
import { ProjectCard } from '@/components/ProjectCard'
import { Pagination } from '@/components/Pagination'
import { PageHeader } from '@/components/PageHeader'
import { EmptyState } from '@/components/EmptyState'
import { SearchX } from 'lucide-react'
import { ProjectSearchForm } from '@/components/talent/ProjectSearchForm'

export const metadata = { title: '案件を探す | Tsunagu Works' }

// searchParams は Next.js 15+ で Promise になった（動的APIの非同期化）
export default async function TalentProjectsPage({
  searchParams,
}: {
  searchParams: Promise<ProjectSearchParams>
}) {
  const params = await searchParams
  const result = await searchProjects(params)

  if (!result) {
    return <p className="text-destructive">案件を取得できませんでした</p>
  }

  const page = currentPage(params)

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="案件を探す"
        description={`${result.total.toLocaleString()}件の募集中案件`}
      />

      <ProjectSearchForm defaultValues={params} />

      {result.projects.length === 0 ? (
        <EmptyState
          icon={SearchX}
          title="条件に合う案件が見つかりませんでした"
          description="条件をゆるめると見つかるかもしれません"
        />
      ) : (
        <ul className="flex flex-col gap-3">
          {result.projects.map((project) => (
            <li key={project.id}>
              <ProjectCard project={project} />
            </li>
          ))}
        </ul>
      )}

      <Pagination
        currentPage={page}
        total={result.total}
        perPage={PROJECTS_PER_PAGE}
      />
    </div>
  )
}
