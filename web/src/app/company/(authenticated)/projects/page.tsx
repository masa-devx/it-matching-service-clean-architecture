import Link from 'next/link'
import { getMyProjects } from '@/lib/projects'
import { Button } from '@/components/ui/button'
import { ProjectCard } from '@/components/ProjectCard'

export const metadata = { title: '案件管理 | Tsunagu Works' }

export default async function CompanyProjectsPage() {
  const projects = await getMyProjects()

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold">案件管理</h1>
          <p className="text-sm text-muted-foreground">
            自社が掲載した案件の一覧です
          </p>
        </div>
        <Button asChild className="h-11">
          <Link href="/company/projects/new">新規掲載</Link>
        </Button>
      </div>

      {projects === null ? (
        <p className="text-destructive">案件を取得できませんでした</p>
      ) : projects.length === 0 ? (
        <p className="rounded-lg border bg-card p-8 text-center text-muted-foreground">
          まだ案件を掲載していません
        </p>
      ) : (
        <ul className="flex flex-col gap-3">
          {projects.map((project) => (
            <li key={project.id}>
              <ProjectCard project={project} />
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
