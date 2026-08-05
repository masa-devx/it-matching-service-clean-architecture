// 「登録したら何が起きるか」が分からないと登録に踏み切れない。
// 手順を示すことで、始めるまでの心理的なハードルを下げる
const steps = [
  {
    title: '会員登録',
    description: '企業／人材を選んでメールアドレスで登録します',
  },
  {
    title: 'プロフィール登録',
    description: '企業は会社情報、人材はスキル・稼働条件・希望単価を入力します',
  },
  {
    title: '掲載・検索',
    description:
      '企業は案件を掲載、人材はスキルや条件で案件を探して詳細を確認します',
  },
  {
    title: '応募・契約',
    description:
      '双方の合意で契約が成立し、稼働報告と検収を経て支払いへ進みます',
  },
]

export function HowItWorks() {
  return (
    <section className="flex flex-col gap-8 px-4 py-16">
      <div className="flex flex-col gap-2 text-center">
        <h2 className="text-2xl font-bold tracking-tight">利用の流れ</h2>
        <p className="text-sm text-muted-foreground">
          登録から契約まで、すべてサービス上で完結します
        </p>
      </div>

      {/* ol を使うのは「順序に意味がある」ことをHTMLで表すため */}
      <ol className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {steps.map(({ title, description }, index) => (
          <li
            key={title}
            className="flex flex-col gap-3 rounded-lg border bg-card p-6"
          >
            <span
              className="flex size-8 items-center justify-center rounded-full bg-primary text-sm font-bold text-primary-foreground"
              aria-hidden="true"
            >
              {index + 1}
            </span>
            <div className="flex flex-col gap-1">
              <h3 className="font-bold">{title}</h3>
              <p className="text-sm leading-relaxed text-muted-foreground">
                {description}
              </p>
            </div>
          </li>
        ))}
      </ol>
    </section>
  )
}
