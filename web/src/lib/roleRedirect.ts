import type { CurrentUser } from './auth'

// ロールごとの着地点。ログイン後・ガードのリダイレクト先をここに一元化する
// （散らばると「片方だけ直し忘れて迷子になる」事故が起きる）
export const dashboardPathByRole: Record<CurrentUser['role'], string> = {
  company: '/company/dashboard',
  talent: '/talent/dashboard',
}

export function dashboardPath(role: CurrentUser['role']): string {
  return dashboardPathByRole[role]
}
