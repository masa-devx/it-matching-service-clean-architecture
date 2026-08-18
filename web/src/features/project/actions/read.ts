'use server'

import { getMyProject, listMyProjects } from '@/external/handler/project'

// TanStack Query の queryFn 用の読み取りアクション（#44 の me と同じ型）。
// クライアントからの再フェッチはここを通ってサーバー側で認証付き取得される。
// id はクライアント供給値だが、所有チェックは API の WHERE company_id が行う（IDOR 対策はサーバーの責務）
export async function fetchMyProjects() {
  return listMyProjects()
}

export async function fetchMyProject(id: number) {
  return getMyProject(id)
}
