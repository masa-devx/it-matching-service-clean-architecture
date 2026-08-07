import Link from 'next/link'
import { Clock, MapPin, Users } from 'lucide-react'
import type { MyProject } from '@/lib/companyProjects'
import { ProjectStatusBadge } from '@/components/company/ProjectStatusBadge'

// 人材向けの ProjectCard とは見せる情報も遷移先も違うため、共通化せず別部品にする
// （人材＝案件を探す / 企業＝応募状況を管理する）

function formatRate(min: number, max: number): string {
  if (min === 0 && max === 0) {
    return '応相談'
  }
  const format = (v: number) => v.toLocaleString('ja-JP')
  return min === max ? `${format(min)}円` : `${format(min)}〜${format(max)}円`
}

export function CompanyProjectCard({ project }: { project: MyProject }) {
  return (
    <Link
      href={`/company/projects/${project.id}/applications`}
      className="flex flex-col gap-3 rounded-lg border bg-card p-4 transition-colors hover:border-primary hover:bg-primary/5"
    >
      <div className="flex flex-wrap items-center gap-2">
        <ProjectStatusBadge status={project.status} />
        <h2 className="font-bold leading-snug">{project.title}</h2>
      </div>

      <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-muted-foreground">
        <span className="tabular-nums">
          {formatRate(project.hourly_rate_min, project.hourly_rate_max)}
        </span>
        {project.hours_per_week > 0 && (
          <span className="flex items-center gap-1.5">
            <Clock className="size-4" aria-hidden="true" />週
            {project.hours_per_week}時間
          </span>
        )}
        {project.remote_ok && (
          <span className="flex items-center gap-1.5">
            <MapPin className="size-4" aria-hidden="true" />
            リモート可
          </span>
        )}
      </div>

      {/* 企業が一覧で最も知りたいのは「応募が来ているか・対応が残っているか」 */}
      <p className="flex items-center gap-1.5 text-sm">
        <Users className="size-4 text-muted-foreground" aria-hidden="true" />
        {project.applications_count === 0 ? (
          <span className="text-muted-foreground">応募はまだありません</span>
        ) : (
          <>
            <span className="font-medium tabular-nums">
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
    </Link>
  )
}
