import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { getCurrentUser } from '@/lib/auth'

export default async function Home() {
  const user = await getCurrentUser()

  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-4">
      <h1 className="text-3xl font-bold tracking-tight">Tsunagu Works</h1>
      <p className="text-muted-foreground">
        企業とIT人材をつなぐビジネスマッチング
      </p>
      <div className="flex gap-3">
        {user ? (
          <Button asChild className="h-11">
            <Link href="/projects">案件一覧へ</Link>
          </Button>
        ) : (
          <>
            <Button asChild className="h-11">
              <Link href="/signup">新規登録</Link>
            </Button>
            <Button asChild variant="outline" className="h-11">
              <Link href="/login">ログイン</Link>
            </Button>
          </>
        )}
      </div>
    </main>
  )
}
