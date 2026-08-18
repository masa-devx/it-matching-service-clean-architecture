import { Features } from '@/components/landing/Features'
import { FinalCta } from '@/components/landing/FinalCta'
import { Hero } from '@/components/landing/Hero'
import { HowItWorks } from '@/components/landing/HowItWorks'
import { ProblemSolution } from '@/components/landing/ProblemSolution'
import { currentRole } from '@/external/handler/auth'
import { dashboardPath } from '@/features/auth/paths'

export const metadata = {
  title: 'Tsunagu Works — 企業とIT人材の安心マッチング',
  description:
    '企業とIT人材（副業・フリーランス）をつなぐビジネスマッチング。エスクロー決済・連絡先マスキング・公平なレビューで安心して取引できます。',
}

export default async function Home() {
  const role = await currentRole()

  return (
    <>
      {/* 幅を持つセクション（背景色を全幅に広げる）は外側、内容は max-w-6xl で中央寄せ */}
      <div className="mx-auto w-full max-w-6xl">
        <Hero dashboardHref={role ? dashboardPath(role) : null} />
        <ProblemSolution />
      </div>
      <Features />
      <div className="mx-auto w-full max-w-6xl">
        <HowItWorks />
      </div>
      {!role && <FinalCta />}
    </>
  )
}
