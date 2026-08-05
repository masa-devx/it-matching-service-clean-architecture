import { Banknote, ShieldOff, Scale, FileClock } from 'lucide-react'
import { Badge } from '@/components/ui/badge'

// 実装状況を正直に示す。ポートフォリオとして「動くもの」と「これから」を
// 区別しないと、見る人が判断を誤る（誠実さは信頼の一部）
const features = [
  {
    icon: Banknote,
    title: 'エスクロー決済',
    description:
      '契約時に企業が仮払い、検収後に人材へ支払い。「未払い」と「前払い」のリスクを同時に解決します。',
    status: 'planned' as const,
  },
  {
    icon: ShieldOff,
    title: '連絡先マスキング',
    description:
      'メッセージ内のメールアドレス・電話番号を検出して伏せ、プラットフォーム外への誘導を抑止します。',
    status: 'planned' as const,
  },
  {
    icon: Scale,
    title: 'レビュー同時公開',
    description:
      '双方が書き終えるまで互いに非公開、そして同時公開。報復レビューを防ぎ評価の信頼性を守ります。',
    status: 'planned' as const,
  },
  {
    icon: FileClock,
    title: '稼働報告',
    description:
      '週次の作業レポートで「働きぶりが見えない」不安と「働いた証拠を残したい」ニーズを同時に解決します。',
    status: 'planned' as const,
  },
]

const statusLabel = {
  available: { label: '利用できます', variant: 'default' as const },
  planned: { label: '開発中', variant: 'outline' as const },
}

export function Features() {
  return (
    <section className="flex flex-col gap-8 bg-card px-4 py-16">
      <div className="mx-auto flex w-full max-w-6xl flex-col gap-8">
        <div className="flex flex-col gap-2 text-center">
          <h2 className="text-2xl font-bold tracking-tight">信頼の設計</h2>
          <p className="text-sm text-muted-foreground">
            マッチングの価値は「出会わせること」ではなく、安心して取引できる仕組みにあると考えています
          </p>
        </div>

        <ul className="grid gap-4 sm:grid-cols-2">
          {features.map(({ icon: Icon, title, description, status }) => (
            <li
              key={title}
              className="flex flex-col gap-3 rounded-lg border bg-background p-6"
            >
              <div className="flex items-start justify-between gap-3">
                <div className="flex items-center gap-3">
                  <Icon className="size-6 text-primary" aria-hidden="true" />
                  <h3 className="font-bold">{title}</h3>
                </div>
                <Badge variant={statusLabel[status].variant}>
                  {statusLabel[status].label}
                </Badge>
              </div>
              <p className="text-sm leading-relaxed text-muted-foreground">
                {description}
              </p>
            </li>
          ))}
        </ul>

        <p className="text-center text-sm text-muted-foreground">
          現在ご利用いただけるのは、会員登録・プロフィール・案件の掲載と検索です
        </p>
      </div>
    </section>
  )
}
