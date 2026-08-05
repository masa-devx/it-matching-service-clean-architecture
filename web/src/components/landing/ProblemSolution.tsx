import { Building2, User } from 'lucide-react'

// 二面市場では、それぞれの立場の課題を「自分ごと」として見せる必要がある。
// 企業と人材を並置し、来訪者が自分の側を見つけられるようにする
const sides = [
  {
    icon: Building2,
    label: '企業（発注側）',
    problems: [
      '正社員を雇う余裕はないが技術力が必要',
      '誰が本当にできる人か見極められない',
      '前払いしたのに音信不通になったら困る',
    ],
    solutions: [
      'スキル・稼働条件で構造化されたプロフィール',
      '検収してから支払うエスクロー決済',
      '週次の稼働報告で進捗が見える',
    ],
  },
  {
    icon: User,
    label: 'IT人材（受注側）',
    problems: [
      '副業を始めたいが最初の案件の入口がない',
      '納品したのに支払われないリスクがある',
      '実績が形に残らない',
    ],
    solutions: [
      'スキル・単価・稼働で絞り込める案件検索',
      '契約時に企業が仮払い済みだから未払いがない',
      '完了した契約とレビューが実績として蓄積',
    ],
  },
]

export function ProblemSolution() {
  return (
    <section className="flex flex-col gap-8 px-4 py-16">
      <div className="flex flex-col gap-2 text-center">
        <h2 className="text-2xl font-bold tracking-tight">
          副業・業務委託の「不安」を仕組みで解決
        </h2>
        <p className="text-sm text-muted-foreground">
          発注側・受注側それぞれの悩みに向き合っています
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {sides.map(({ icon: Icon, label, problems, solutions }) => (
          <div
            key={label}
            className="flex flex-col gap-6 rounded-lg border bg-card p-6"
          >
            <div className="flex items-center gap-3">
              <Icon className="size-6 text-primary" aria-hidden="true" />
              <h3 className="font-bold">{label}</h3>
            </div>

            <div className="flex flex-col gap-3">
              <p className="text-sm font-medium text-muted-foreground">
                こんな悩み
              </p>
              <ul className="flex flex-col gap-2">
                {problems.map((problem) => (
                  <li key={problem} className="text-sm">
                    「{problem}」
                  </li>
                ))}
              </ul>
            </div>

            <div className="flex flex-col gap-3">
              <p className="text-sm font-medium text-primary">解決する仕組み</p>
              <ul className="flex flex-col gap-2">
                {solutions.map((solution) => (
                  <li key={solution} className="flex gap-2 text-sm">
                    <span aria-hidden="true">→</span>
                    <span>{solution}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
