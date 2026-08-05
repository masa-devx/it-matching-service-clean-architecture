import { getCurrentUser } from '@/lib/auth'
import { Hero } from '@/components/landing/Hero'
import { ProblemSolution } from '@/components/landing/ProblemSolution'
import { Features } from '@/components/landing/Features'
import { HowItWorks } from '@/components/landing/HowItWorks'
import { FinalCta } from '@/components/landing/FinalCta'

export const metadata = {
  title: 'Tsunagu Works — 企業とIT人材の安心マッチング',
  description:
    '企業とIT人材（副業・フリーランス）をつなぐビジネスマッチング。エスクロー決済・連絡先マスキング・公平なレビューで安心して取引できます。',
}

export default async function Home() {
  const user = await getCurrentUser()

  return (
    <>
      {/* 幅を持つセクション（背景色を全幅に広げる）は外側、内容は max-w-6xl で中央寄せ */}
      <div className="mx-auto w-full max-w-6xl">
        <Hero user={user} />
        <ProblemSolution />
      </div>
      <Features />
      <div className="mx-auto w-full max-w-6xl">
        <HowItWorks />
      </div>
      {!user && <FinalCta />}
    </>
  )
}
