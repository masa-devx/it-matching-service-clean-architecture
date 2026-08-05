import Link from 'next/link'
import type { CurrentUser } from '@/lib/auth'
import { dashboardPath } from '@/lib/roleRedirect'
import { Button } from '@/components/ui/button'

// LP用の簡易ヘッダー。業務画面の AppHeader とは役割が違う
// （こちらは「サービスを知る/入口へ進む」ためのもので、ナビゲーションは持たない）
export function LandingHeader({ user }: { user: CurrentUser | null }) {
  return (
    <header className="sticky top-0 z-10 border-b bg-card/80 backdrop-blur">
      <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between px-4">
        <Link href="/" className="text-lg font-bold text-primary">
          Tsunagu Works
        </Link>
        {user ? (
          <Button asChild className="h-11">
            <Link href={dashboardPath(user.role)}>ダッシュボードへ</Link>
          </Button>
        ) : (
          <div className="flex items-center gap-2">
            <Button asChild variant="ghost" className="h-11">
              <Link href="/login">ログイン</Link>
            </Button>
            <Button asChild className="h-11">
              <Link href="/signup">無料で始める</Link>
            </Button>
          </div>
        )}
      </div>
    </header>
  )
}
