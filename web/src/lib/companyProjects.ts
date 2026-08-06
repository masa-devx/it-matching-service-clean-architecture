import 'server-only'

import { getMyProjects, type Project } from './projects'
import { getProjectApplications } from './companyApplications'

// 案件＋応募の件数。一覧で「どの案件に応募が来ているか」を判断するために使う
export type ProjectWithApplications = Project & {
  applicationCount: number
  // まだ企業が対応していない応募（applied）。ここが「次にやること」を示す
  pendingCount: number
}

// 自社案件に応募件数を添えて返す。
//
// 案件一覧APIが件数を返さないため、案件ごとに応募一覧APIを叩いている（fan-out）。
// limit=1 で total だけ受け取り、Promise.all で並列化することで転送量と待ち時間を抑えている。
// 掲載案件が増えるとリクエスト数も比例して増えるため、
// 本来は案件一覧APIが applications_count を返すべき（#93 で対応する）
export async function getMyProjectsWithApplications(): Promise<
  ProjectWithApplications[] | null
> {
  const projects = await getMyProjects()
  if (!projects) {
    return null
  }

  return Promise.all(
    projects.map(async (project) => {
      const [all, pending] = await Promise.all([
        getProjectApplications(project.id, {}, 1),
        getProjectApplications(project.id, { status: 'applied' }, 1),
      ])
      return {
        ...project,
        applicationCount: all?.total ?? 0,
        pendingCount: pending?.total ?? 0,
      }
    }),
  )
}
