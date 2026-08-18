import { SkipLink } from '@/components/SkipLink'
import { LandingFooter } from '@/components/landing/LandingFooter'
import { LandingHeader } from '@/components/landing/LandingHeader'
import { currentRole } from '@/external/handler/auth'
import { dashboardPath } from '@/features/auth/paths'

// (public) は誰でも閲覧できる。ログイン状態はガードには使わず、
// 「ダッシュボードへ」の導線を出すかどうかの判定にだけ使う。
// currentRole は cache() 済みなので page 側と合わせても API 呼び出しは1回
export default async function PublicLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const role = await currentRole()

  return (
    <div className="flex flex-1 flex-col">
      <SkipLink />
      <LandingHeader dashboardHref={role ? dashboardPath(role) : null} />
      <main id="main" className="flex-1">
        {children}
      </main>
      <LandingFooter />
    </div>
  )
}
