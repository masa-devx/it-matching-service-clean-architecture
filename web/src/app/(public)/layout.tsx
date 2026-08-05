import { getCurrentUser } from '@/lib/auth'
import { LandingHeader } from '@/components/landing/LandingHeader'
import { LandingFooter } from '@/components/landing/LandingFooter'

// (public) は誰でも閲覧できる。ログイン状態はガードには使わず、
// 「ダッシュボードへ」の導線を出すかどうかの判定にだけ使う
export default async function PublicLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const user = await getCurrentUser()

  return (
    <div className="flex flex-1 flex-col">
      <LandingHeader user={user} />
      <main className="flex-1">{children}</main>
      <LandingFooter />
    </div>
  )
}
