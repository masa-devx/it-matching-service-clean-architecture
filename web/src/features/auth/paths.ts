// ロールごとの着地点。ログイン後・ガード・LP導線のリダイレクト先をここに一元化する
// （散らばると「片方だけ直し忘れて迷子になる」事故が起きる）

export type Role = 'company' | 'talent'

export const dashboardPathByRole: Record<Role, string> = {
  company: '/company/dashboard',
  talent: '/talent/dashboard',
}

export function dashboardPath(role: Role): string {
  return dashboardPathByRole[role]
}
