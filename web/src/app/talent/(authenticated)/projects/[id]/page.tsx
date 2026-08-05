import Link from 'next/link'
import { notFound } from 'next/navigation'
import { ArrowLeft, Banknote, Clock, MapPin } from 'lucide-react'
import { getProject } from '@/lib/projects'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'

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

  // 条件は「単価・稼働・勤務形態」の3点セットで判断されるため、まとめて上部に置く
  const conditions = [
    {
      icon: Banknote,
      label: '時給',
      value: formatRate(project.hourly_rate_min, project.hourly_rate_max),
      emphasized: project.hourly_rate_min > 0 || project.hourly_rate_max > 0,
    },
    {
      icon: Clock,
      label: '週の稼働時間',
      value:
        project.hours_per_week > 0 ? `${project.hours_per_week}時間` : '応相談',
      emphasized: project.hours_per_week > 0,
    },
    {
      icon: MapPin,
      label: '勤務形態',
      value: project.remote_ok ? 'フルリモート可' : '出社あり',
      emphasized: project.remote_ok,
    },
  ]

  return (
    <article className="flex flex-col gap-6">
      <Link
        href="/talent/projects"
        className="flex w-fit items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="size-4" aria-hidden="true" />
        案件一覧に戻る
      </Link>

      {/* 見出しと会社名。詳細では PageHeader を使わず、条件カードと視覚的に繋げる */}
      <header className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold leading-snug tracking-tight">
          {project.title}
        </h1>
        <p className="text-muted-foreground">{project.company_name}</p>
      </header>

      {/* 応募判断に直結する条件を最上部にまとめる（スクロールせずに見える位置） */}
      <dl className="grid gap-4 rounded-lg border bg-card p-6 sm:grid-cols-3">
        {conditions.map(({ icon: Icon, label, value, emphasized }) => (
          <div key={label} className="flex items-start gap-3">
            <Icon className="mt-0.5 size-5 text-primary" aria-hidden="true" />
            <div className="flex flex-col gap-0.5">
              <dt className="text-sm text-muted-foreground">{label}</dt>
              <dd
                className={`font-bold tabular-nums ${emphasized ? '' : 'text-muted-foreground'}`}
              >
                {value}
              </dd>
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
        <>
          <Separator />
          <section className="flex flex-col gap-3">
            <h2 className="font-bold">案件内容</h2>
            {/* 改行を保持しつつ、長い単語で横スクロールしないようにする */}
            <p className="whitespace-pre-wrap break-words leading-relaxed">
              {project.description}
            </p>
          </section>
        </>
      )}

      <Separator />

      {/* 応募機能は Phase 3（状態機械①）で実装する */}
      <div className="flex flex-col gap-2">
        <Button className="h-11 self-start" disabled>
          応募する
        </Button>
        <p className="text-sm text-muted-foreground">
          応募機能は現在開発中です
        </p>
      </div>
    </article>
  )
}
