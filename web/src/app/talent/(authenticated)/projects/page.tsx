import {
  searchProjects,
  currentPage,
  PROJECTS_PER_PAGE,
  type ProjectSearchParams,
} from '@/lib/projects'
import { ProjectCard } from '@/components/ProjectCard'
import { Pagination } from '@/components/Pagination'
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
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold">案件を探す</h1>
        <p className="text-sm text-muted-foreground">
          <span className="tabular-nums">{result.total.toLocaleString()}</span>
          件の募集中案件
        </p>
      </div>

      <ProjectSearchForm defaultValues={params} />

      {result.projects.length === 0 ? (
        <p className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
          条件に合う案件が見つかりませんでした
        </p>
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
