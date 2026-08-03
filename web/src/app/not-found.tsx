import Link from 'next/link'
import { Button } from '@/components/ui/button'

export const metadata = { title: 'ページが見つかりません | Tsunagu Works' }

// 存在しないURL・notFound() 呼び出し時に表示される（アプリ全体の404）
export default function NotFound() {
  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-4 py-16">
      <p className="text-sm font-medium text-muted-foreground">404</p>
      <h1 className="text-2xl font-bold">ページが見つかりません</h1>
      <p className="text-sm text-muted-foreground">
        URLが変更されたか、削除された可能性があります。
      </p>
      <Button asChild className="h-11">
        <Link href="/">トップへ戻る</Link>
      </Button>
    </main>
  )
}
