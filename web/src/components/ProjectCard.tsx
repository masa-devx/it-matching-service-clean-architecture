import Link from 'next/link'
import type { Project } from '@/lib/projects'

// 時給は「4,000〜6,000円」の形に整形する。未設定（0）は表示しない
function formatRate(min: number, max: number): string {
  if (min === 0 && max === 0) {
    return '応相談'
  }
  const format = (v: number) => v.toLocaleString('ja-JP')
  return min === max ? `${format(min)}円` : `${format(min)}〜${format(max)}円`
}

export function ProjectCard({ project }: { project: Project }) {
  return (
    <Link
      href={`/talent/projects/${project.id}`}
      className="flex flex-col gap-3 rounded-lg border bg-card p-4 transition-colors hover:border-primary"
    >
      <div className="flex flex-col gap-1">
        <h2 className="font-bold">{project.title}</h2>
        <p className="text-sm text-muted-foreground">{project.company_name}</p>
      </div>

      {project.required_skills.length > 0 && (
        <ul className="flex flex-wrap gap-1.5">
          {project.required_skills.map((skill) => (
            <li
              key={skill}
              className="rounded-full bg-secondary px-2.5 py-0.5 text-xs"
            >
              {skill}
            </li>
          ))}
        </ul>
      )}

      <dl className="flex flex-wrap gap-x-6 gap-y-1 text-sm">
        <div className="flex gap-1.5">
          <dt className="text-muted-foreground">時給</dt>
          {/* tabular-nums: 桁が変わっても数字幅が揃い、一覧の見た目が安定する */}
          <dd className="font-medium tabular-nums">
            {formatRate(project.hourly_rate_min, project.hourly_rate_max)}
          </dd>
        </div>
        {project.hours_per_week > 0 && (
          <div className="flex gap-1.5">
            <dt className="text-muted-foreground">稼働</dt>
            <dd className="font-medium tabular-nums">
              週{project.hours_per_week}時間
            </dd>
          </div>
        )}
        {project.remote_ok && (
          <div className="flex gap-1.5">
            <dd className="font-medium text-primary">フルリモート可</dd>
          </div>
        )}
      </dl>
    </Link>
  )
}
