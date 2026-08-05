import Link from 'next/link'
import { Separator } from '@/components/ui/separator'

// 技術スタックを明示するのは、このリポジトリが転職ポートフォリオでもあるため。
// 採用担当が「何で作られているか」をすぐ確認できる導線にする
const techStack = [
  'Go (net/http)',
  'PostgreSQL',
  'Next.js (App Router)',
  'TypeScript',
  'Tailwind CSS',
  'Docker',
]

export function LandingFooter() {
  return (
    <footer className="border-t bg-card">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-6 px-4 py-10">
        <div className="flex flex-col gap-2">
          <p className="font-bold text-primary">Tsunagu Works</p>
          <p className="text-sm text-muted-foreground">
            企業とIT人材が安心して取引できるビジネスマッチング
          </p>
        </div>

        <Separator />

        <div className="flex flex-col gap-3">
          <p className="text-sm font-medium">技術スタック</p>
          <ul className="flex flex-wrap gap-x-4 gap-y-1 text-sm text-muted-foreground">
            {techStack.map((tech) => (
              <li key={tech}>{tech}</li>
            ))}
          </ul>
        </div>

        <div className="flex flex-wrap items-center justify-between gap-4 text-sm text-muted-foreground">
          <Link
            href="https://github.com/masahiro96848/tsunagu-works"
            className="underline hover:text-foreground"
            target="_blank"
            rel="noopener noreferrer"
          >
            GitHub でソースコードを見る
          </Link>
          <p>個人開発のポートフォリオとして制作しています</p>
        </div>
      </div>
    </footer>
  )
}
