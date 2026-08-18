import { notFound } from 'next/navigation'

import { ProjectForm } from '@/features/project/components/client/ProjectForm'
import { getMyProject } from '@/external/handler/project'

// Next.js 15+ では params は Promise（await が必須）。
// 編集フォームの初期値はサーバーで取得して渡す（キャッシュ不要のワンショット読み）。
// 他社の案件・不存在の id はどちらも handler が失敗を返すので同じ 404 に落とす
export default async function Page({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const projectId = Number(id)
  if (!Number.isInteger(projectId) || projectId <= 0) {
    notFound()
  }

  const result = await getMyProject(projectId)
  if (!result.ok) {
    notFound()
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">案件を編集</h1>
      <ProjectForm project={result.data} />
    </div>
  )
}
