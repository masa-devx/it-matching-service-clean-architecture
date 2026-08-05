import { RoleChoice } from '@/components/RoleChoice'

export const metadata = { title: 'ログイン | Tsunagu Works' }

// 企業/人材の入口を選ぶ画面。実際のログインフォームは各ロール配下にある
export default function LoginChoicePage() {
  return <RoleChoice action="login" />
}
