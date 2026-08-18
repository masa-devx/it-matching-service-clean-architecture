import { Badge } from '@/components/ui/badge'

// スキル配列の表示（4画面で共用）。ドメイン知識ゼロ・props だけの表示部品なので
// feature 複製ではなく components/ に置く（rateText を視点ごとに複製した判断との対比:
// あちらは「視点で育つ表示ロジック」、こちらは「どの視点でも同じ視覚語彙」）
export function SkillBadges({ skills }: { skills: string[] }) {
  if (skills.length === 0) {
    return null
  }
  return (
    <div className="flex flex-wrap gap-1.5">
      {skills.map((skill) => (
        <Badge key={skill} variant="secondary">
          {skill}
        </Badge>
      ))}
    </div>
  )
}
