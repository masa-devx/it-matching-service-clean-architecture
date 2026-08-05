import Link from 'next/link'
import { notFound } from 'next/navigation'
import { getProject } from '@/lib/projects'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { PageHeader } from '@/components/PageHeader'

// 時給・稼働の表示は一覧と揃える（未設定=0 は「応相談」）
function formatRate(min: number, max: number): string {
  if (min === 0 && max === 0) {
    return '応相談'
  }
  const format = (v: number) => v.toLocaleString('ja-JP')
  return min === max ? `${format(min)}円` : `${format(min)}〜${format(max)}円`
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const project = await getProject(Number(id))
  return {
    title: `${project?.title ?? '案件が見つかりません'} | Tsunagu Works`,
  }
}

export default async function ProjectDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const project = await getProject(Number(id))
  // 存在しない・未公開・不正なIDはすべて 404（下書きの存在を外部に漏らさない）
  if (!project) {
    notFound()
  }

  return (
    <article className="flex flex-col gap-6">
      <Link
        href="/talent/projects"
        className="text-sm text-muted-foreground hover:text-foreground"
      >
        ← 案件一覧に戻る
      </Link>

      <PageHeader title={project.title} description={project.company_name} />

      <dl className="grid gap-4 rounded-lg border bg-card p-4 sm:grid-cols-3">
        <div className="flex flex-col gap-1">
          <dt className="text-sm text-muted-foreground">時給</dt>
          <dd className="font-medium tabular-nums">
            {formatRate(project.hourly_rate_min, project.hourly_rate_max)}
          </dd>
        </div>
        <div className="flex flex-col gap-1">
          <dt className="text-sm text-muted-foreground">週の稼働時間</dt>
          <dd className="font-medium tabular-nums">
            {project.hours_per_week > 0
              ? `${project.hours_per_week}時間`
              : '応相談'}
          </dd>
        </div>
        <div className="flex flex-col gap-1">
          <dt className="text-sm text-muted-foreground">勤務形態</dt>
          <dd className="font-medium">
            {project.remote_ok ? 'フルリモート可' : '出社あり'}
          </dd>
        </div>
      </dl>

      {project.required_skills.length > 0 && (
        <section className="flex flex-col gap-2">
          <h2 className="font-bold">必須スキル</h2>
          <ul className="flex flex-wrap gap-2">
            {project.required_skills.map((skill) => (
              <li key={skill}>
                <Badge variant="secondary">{skill}</Badge>
              </li>
            ))}
          </ul>
        </section>
      )}

      {project.description && (
        <section className="flex flex-col gap-2">
          <h2 className="font-bold">案件内容</h2>
          {/* 改行を保持しつつ、長い単語で横スクロールしないようにする */}
          <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">
            {project.description}
          </p>
        </section>
      )}

      {/* 応募機能は Phase 3（状態機械①）で実装する */}
      <Button className="h-11 self-start" disabled>
        応募する（準備中）
      </Button>
    </article>
  )
}
