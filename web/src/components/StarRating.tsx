import { Star } from 'lucide-react'
import { RATING_VALUES } from '@/lib/reviewSchema'

// 評価の表示（読み取り専用）。
//
// 星の形だけでは、色覚特性のある人や読み上げ環境で伝わらない。
// aria-label で「5段階中4」と読ませ、数値も併記する
// （#75 の「色だけで意味を伝えない」の適用）
export function StarRating({
  rating,
  label,
}: {
  rating: number
  // 補足の文言（「期待を上回った」など）。基準を言葉で示して解釈を揃える
  label?: string
}) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      <span
        className="flex items-center gap-0.5"
        role="img"
        aria-label={`5段階中${rating}`}
      >
        {RATING_VALUES.map((value) => (
          <Star
            key={value}
            className={`size-4 ${
              value <= rating
                ? 'fill-primary text-primary'
                : 'text-muted-foreground/40'
            }`}
            aria-hidden="true"
          />
        ))}
      </span>
      <span className="text-sm font-medium tabular-nums">{rating}</span>
      {label && (
        <span className="text-sm text-muted-foreground">（{label}）</span>
      )}
    </div>
  )
}
