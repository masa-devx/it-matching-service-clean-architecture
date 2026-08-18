'use server'

import type { ProjectsListParams } from '@repo/api-client/talent/generated/models'

import {
  getPublishedProject,
  searchProjects,
} from '@/external/handler/projectSearch'

// queryFn の通り道（#44/#45 と同じ型）。検索条件・カーソルはクライアント供給値だが、
// 見えるのは公開案件だけという境界は SQL（WHERE status='published'）が保証する
export async function fetchProjects(params: ProjectsListParams) {
  return searchProjects(params)
}

export async function fetchProject(id: number) {
  return getPublishedProject(id)
}
