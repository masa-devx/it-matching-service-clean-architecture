import Link from 'next/link'

import { Button } from '@/components/ui/button'

// LP用の簡易ヘッダー。業務画面のヘッダーとは役割が違う
// （こちらは「サービスを知る/入口へ進む」ためのもので、ナビゲーションは持たない）。
// 認証状態は知らず、app 側が計算した遷移先（dashboardHref）だけを受け取る
export function LandingHeader({
  dashboardHref,
}: {
  dashboardHref: string | null
}) {
  return (
    <header className="sticky top-0 z-10 border-b bg-card/80 backdrop-blur">
      <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between px-4">
        <Link href="/" className="text-lg font-bold text-primary">
          Tsunagu Works
        </Link>
        {dashboardHref ? (
          <Button asChild className="h-11">
            <Link href={dashboardHref}>ダッシュボードへ</Link>
          </Button>
        ) : (
          <div className="flex items-center gap-2">
            <Button asChild variant="ghost" className="h-11">
              <Link href="/company/login">企業ログイン</Link>
            </Button>
            <Button asChild variant="ghost" className="h-11">
              <Link href="/talent/login">人材ログイン</Link>
            </Button>
            <Button asChild className="h-11">
              <Link href="/#cta">無料で始める</Link>
            </Button>
          </div>
        )}
      </div>
    </header>
  )
}
