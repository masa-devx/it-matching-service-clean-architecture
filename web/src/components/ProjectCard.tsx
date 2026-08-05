import Link from 'next/link'
import { Clock, MapPin } from 'lucide-react'
import type { Project } from '@/lib/projects'
import { Badge } from '@/components/ui/badge'

// 案件探しで最初に見るのは「いくらか」「どれくらい働くか」。
// タイトル → 単価 → 条件 の順に視覚的な重みを付ける
function formatRate(min: number, max: number): string {
  if (min === 0 && max === 0) {
    return '応相談'
  }
  const format = (v: number) => v.toLocaleString('ja-JP')
  return min === max ? `${format(min)}円` : `${format(min)}〜${format(max)}円`
}

// 3日以内の掲載を「新着」として目立たせる（探す側にとって鮮度は重要な情報）
const NEW_THRESHOLD_DAYS = 3

function isNew(createdAt: string): boolean {
  const elapsed = Date.now() - new Date(createdAt).getTime()
  return elapsed < NEW_THRESHOLD_DAYS * 24 * 60 * 60 * 1000
}

export function ProjectCard({ project }: { project: Project }) {
  const hasRate = project.hourly_rate_min > 0 || project.hourly_rate_max > 0

  return (
    <Link
      href={`/talent/projects/${project.id}`}
      className="flex flex-col gap-3 rounded-lg border bg-card p-4 transition-colors hover:border-primary hover:bg-primary/5"
    >
      <div className="flex flex-col gap-1">
        <div className="flex flex-wrap items-center gap-2">
          {isNew(project.created_at) && <Badge>新着</Badge>}
          <h2 className="font-bold leading-snug">{project.title}</h2>
        </div>
        <p className="text-sm text-muted-foreground">{project.company_name}</p>
      </div>

      {/* 単価は探す側の判断材料の中心なので、サイズと色で主役にする */}
      <div className="flex flex-wrap items-baseline gap-x-4 gap-y-1">
        <p
          className={`text-lg font-bold tabular-nums ${hasRate ? 'text-primary' : 'text-muted-foreground'}`}
        >
          {formatRate(project.hourly_rate_min, project.hourly_rate_max)}
          {hasRate && <span className="text-sm font-medium"> / 時</span>}
        </p>

        <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-muted-foreground">
          {project.hours_per_week > 0 && (
            <span className="flex items-center gap-1">
              <Clock className="size-4" aria-hidden="true" />
              <span className="tabular-nums">
                週{project.hours_per_week}時間
              </span>
            </span>
          )}
          {project.remote_ok && (
            <span className="flex items-center gap-1">
              <MapPin className="size-4" aria-hidden="true" />
              フルリモート可
            </span>
          )}
        </div>
      </div>

      {project.required_skills.length > 0 && (
        <ul className="flex flex-wrap gap-2">
          {project.required_skills.map((skill) => (
            <li key={skill}>
              <Badge variant="secondary">{skill}</Badge>
            </li>
          ))}
        </ul>
      )}
    </Link>
  )
}
