import Link from 'next/link'
import { Button } from '@/components/ui/button'

// ページ末尾でもう一度入口を示す。
// 上まで戻らせない（スクロールし切った人が最も関心が高い）
export function FinalCta() {
  return (
    <section
      id="cta"
      className="flex flex-col items-center gap-6 bg-card px-4 py-16 text-center"
    >
      <div className="flex max-w-xl flex-col gap-2">
        <h2 className="text-2xl font-bold tracking-tight">
          まずは無料で始めてみませんか
        </h2>
        <p className="text-sm text-muted-foreground">
          登録は1分。掲載・検索は無料でご利用いただけます
        </p>
      </div>

      <div className="flex flex-wrap justify-center gap-3">
        <Button asChild className="h-11">
          <Link href="/company/signup">企業として登録する</Link>
        </Button>
        <Button asChild variant="outline" className="h-11">
          <Link href="/talent/signup">人材として登録する</Link>
        </Button>
      </div>
    </section>
  )
}
