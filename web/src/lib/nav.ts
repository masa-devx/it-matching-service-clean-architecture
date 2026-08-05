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

// 認証画面でロールごとに出し分ける文言。
// 「どちら向けの画面か」は見た目より言葉で伝わるため、1か所に集約して揺れを防ぐ
export const authCopyByRole: Record<
  CurrentUser['role'],
  { signupTitle: string; signupDescription: string; loginTitle: string }
> = {
  company: {
    signupTitle: '企業として登録',
    signupDescription: '案件を掲載して、必要なスキルを持つ人材を探せます',
    loginTitle: '企業ログイン',
  },
  talent: {
    signupTitle: '人材として登録',
    signupDescription: 'スキルや稼働条件に合う案件を探して応募できます',
    loginTitle: '人材ログイン',
  },
}
