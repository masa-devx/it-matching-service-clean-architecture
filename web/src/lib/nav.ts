import type { CurrentUser } from './auth'

export type NavItem = { href: string; label: string }

// ロールごとのナビ定義。ヘッダー以外（ダッシュボードの導線など）からも参照できるよう
// データとして持ち、描画側は map するだけにする
export const navItemsByRole: Record<CurrentUser['role'], NavItem[]> = {
  company: [
    { href: '/company/dashboard', label: 'ダッシュボード' },
    { href: '/company/projects', label: '案件管理' },
    { href: '/company/profile', label: 'プロフィール' },
  ],
  talent: [
    { href: '/talent/dashboard', label: 'ダッシュボード' },
    { href: '/talent/projects', label: '案件を探す' },
    { href: '/talent/profile', label: 'プロフィール' },
  ],
}

export const roleLabel: Record<CurrentUser['role'], string> = {
  company: '企業',
  talent: '人材',
}
