import Link from 'next/link'
import { notFound } from 'next/navigation'
import { ArrowLeft } from 'lucide-react'
import { getMyProject } from '@/lib/companyProjects'
import { ProjectForm } from '@/components/company/ProjectForm'
import { PageHeader } from '@/components/PageHeader'

export const metadata = { title: '案件を編集 | Tsunagu Works' }

export default async function EditProjectPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const project = await getMyProject(Number(id))
  // 他社の案件・存在しない案件は API が404を返す
  if (!project) {
    notFound()
  }

  return (
    <div className="flex flex-col gap-6">
      <Link
        href={`/company/projects/${project.id}`}
        className="flex w-fit items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="size-4" aria-hidden="true" />
        案件の詳細に戻る
      </Link>

      <PageHeader
        title="案件を編集する"
        description="掲載状態は変わりません（公開・終了は詳細画面から操作します）"
      />

      <ProjectForm project={project} />
    </div>
  )
}
