import Link from 'next/link'
import { notFound } from 'next/navigation'
import { ArrowLeft, Banknote, Clock, MapPin, Users } from 'lucide-react'
import { getMyProject } from '@/lib/companyProjects'
import {
  ProjectStatusBadge,
  projectStatusDescription,
} from '@/components/company/ProjectStatusBadge'
import { ProjectStatusActions } from '@/components/company/ProjectStatusActions'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'

export const metadata = { title: '案件の詳細 | Tsunagu Works' }

function formatRate(min: number, max: number): string {
  if (min === 0 && max === 0) {
    return '応相談'
  }
  const format = (v: number) => v.toLocaleString('ja-JP')
  return min === max ? `${format(min)}円` : `${format(min)}〜${format(max)}円`
}

export default async function CompanyProjectDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const project = await getMyProject(Number(id))
  // 他社の案件・存在しない案件は API が404を返す（存在を漏らさない方針）
  if (!project) {
    notFound()
  }

  const conditions = [
    {
      icon: Banknote,
      label: '時給',
      value: formatRate(project.hourly_rate_min, project.hourly_rate_max),
    },
    {
      icon: Clock,
      label: '週の稼働時間',
      value:
        project.hours_per_week > 0 ? `${project.hours_per_week}時間` : '応相談',
    },
    {
      icon: MapPin,
      label: '勤務形態',
      value: project.remote_ok ? 'フルリモート可' : '出社あり',
    },
  ]

  return (
    <article className="flex flex-col gap-6">
      <Link
        href="/company/projects"
        className="flex w-fit items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="size-4" aria-hidden="true" />
        案件管理に戻る
      </Link>

      <header className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center gap-2">
          <ProjectStatusBadge status={project.status} />
          <span className="text-sm text-muted-foreground">
            {projectStatusDescription(project.status)}
          </span>
        </div>
        <h1 className="text-2xl font-bold leading-snug tracking-tight">
          {project.title}
        </h1>
      </header>

      {/* 掲載状態の操作。現在の状態で実行できるものだけが並ぶ（遷移表の写し） */}
      <section className="flex flex-col gap-3 rounded-lg border bg-card p-6">
        <h2 className="font-bold">掲載状態</h2>
        <ProjectStatusActions projectId={project.id} status={project.status} />
      </section>

      {/* 応募状況。下書きのうちは応募が来ないため、公開後に意味を持つ */}
      <section className="flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-card p-6">
        <p className="flex items-center gap-2">
          <Users className="size-5 text-muted-foreground" aria-hidden="true" />
          {project.applications_count === 0 ? (
            <span className="text-muted-foreground">応募はまだありません</span>
          ) : (
            <>
              <span className="font-bold tabular-nums">
                応募 {project.applications_count}件
              </span>
              {project.pending_count > 0 && (
                <span className="font-medium text-primary tabular-nums">
                  （{project.pending_count}件が未対応）
                </span>
              )}
            </>
          )}
        </p>
        <Button asChild variant="outline" className="h-11">
          <Link href={`/company/projects/${project.id}/applications`}>
            応募者を見る
          </Link>
        </Button>
      </section>

      <Separator />

      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="font-bold">掲載内容</h2>
        <Button asChild variant="outline" className="h-11">
          <Link href={`/company/projects/${project.id}/edit`}>編集する</Link>
        </Button>
      </div>

      <dl className="grid gap-4 rounded-lg border bg-card p-6 sm:grid-cols-3">
        {conditions.map(({ icon: Icon, label, value }) => (
          <div key={label} className="flex items-start gap-3">
            <Icon className="mt-0.5 size-5 text-primary" aria-hidden="true" />
            <div className="flex flex-col gap-0.5">
              <dt className="text-sm text-muted-foreground">{label}</dt>
              <dd className="font-bold tabular-nums">{value}</dd>
            </div>
          </div>
        ))}
      </dl>

      {project.required_skills.length > 0 && (
        <section className="flex flex-col gap-3">
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
        <section className="flex flex-col gap-3">
          <h2 className="font-bold">案件内容</h2>
          <p className="whitespace-pre-wrap break-words leading-relaxed">
            {project.description}
          </p>
        </section>
      )}
    </article>
  )
}
