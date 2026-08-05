// ダッシュボードの数値カード。値の意味（ラベル）と数値の対を統一した見た目で並べる
export function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div className="flex flex-col gap-1 rounded-lg border bg-card p-4">
      <span className="text-sm text-muted-foreground">{label}</span>
      {/* tabular-nums: 桁が変わっても数字の幅が揃い、レイアウトが揺れない */}
      <span className="text-2xl font-bold tabular-nums">{value}</span>
    </div>
  )
}
