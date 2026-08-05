import { getCurrentUser } from '@/lib/auth'
import { Hero } from '@/components/landing/Hero'

export const metadata = {
  title: 'Tsunagu Works — 企業とIT人材の安心マッチング',
  description:
    '企業とIT人材（副業・フリーランス）をつなぐビジネスマッチング。エスクロー決済・連絡先マスキング・公平なレビューで安心して取引できます。',
}

export default async function Home() {
  const user = await getCurrentUser()

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col">
      <Hero user={user} />
    </div>
  )
}
