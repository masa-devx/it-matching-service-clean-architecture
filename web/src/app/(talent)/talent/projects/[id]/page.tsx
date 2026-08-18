import Link from 'next/link'
import { notFound } from 'next/navigation'

import { Button } from '@/components/ui/button'
import { listMyApplications } from '@/external/handler/application'
import { getPublishedProject } from '@/external/handler/projectSearch'
import { ApplicationStatusBadge } from '@/features/application/components/ApplicationStatusBadge'
import { ApplyForm } from '@/features/application/components/client/ApplyForm'

// 詳細はワンショット読みなので handler 直呼び（#45 編集ページと同型）。
// 未公開（draft / closed）と不存在は API がどちらも 404 を返すため、同じ notFound() に落ちる
// ＝ id 総当たりでも未公開案件の存在は分からない
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

  const result = await getPublishedProject(projectId)
  if (!result.ok) {
    notFound()
  }
  const project = result.data

  // 応募済みならフォームの代わりに状態を出す（UIでも二重応募を防ぐ。保証は DB の UNIQUE）
  const applications = await listMyApplications()
  const myApplication = applications.ok
    ? applications.data.find((a) => a.project_id === projectId)
    : undefined

  const rate =
    project.hourly_rate_min == null && project.hourly_rate_max == null
      ? '未設定'
      : `${project.hourly_rate_min ?? ''}円 〜 ${project.hourly_rate_max ?? ''}円`

  return (
    <div className="space-y-6">
      <div>
        <Button asChild variant="ghost" size="sm">
          <Link href="/talent/projects">← 案件一覧へ</Link>
        </Button>
      </div>

      <div className="space-y-1">
        <h1 className="text-2xl font-bold">{project.title}</h1>
        <p className="text-sm text-muted-foreground">
          掲載日: {new Date(project.created_at).toLocaleDateString('ja-JP')}
        </p>
      </div>

      <dl className="space-y-2 rounded-lg border p-4">
        <div>
          <dt className="text-sm text-muted-foreground">想定時給</dt>
          <dd className="font-medium">{rate}</dd>
        </div>
        <div>
          <dt className="text-sm text-muted-foreground">週の稼働時間</dt>
          <dd className="font-medium">週{project.hours_per_week}時間</dd>
        </div>
        <div>
          <dt className="text-sm text-muted-foreground">リモート</dt>
          <dd className="font-medium">{project.remote_ok ? '可' : '不可'}</dd>
        </div>
        {project.required_skills.length > 0 && (
          <div>
            <dt className="text-sm text-muted-foreground">必須スキル</dt>
            <dd className="font-medium">
              {project.required_skills.join(' / ')}
            </dd>
          </div>
        )}
      </dl>

      <section className="space-y-2">
        <h2 className="text-lg font-bold">案件詳細</h2>
        <p className="whitespace-pre-wrap text-sm leading-relaxed">
          {project.description}
        </p>
      </section>

      <section className="space-y-2">
        <h2 className="text-lg font-bold">応募</h2>
        {myApplication ? (
          <div className="flex items-center justify-between gap-2 rounded-lg border p-4">
            <div className="flex items-center gap-2">
              <span className="text-sm">この案件には応募済みです</span>
              <ApplicationStatusBadge status={myApplication.status} />
            </div>
            <Button asChild variant="outline" size="sm">
              <Link href="/talent/applications">応募一覧へ</Link>
            </Button>
          </div>
        ) : (
          <ApplyForm projectId={projectId} />
        )}
      </section>
    </div>
  )
}
