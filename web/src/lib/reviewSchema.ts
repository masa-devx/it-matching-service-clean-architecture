import { z } from 'zod'

// Go 側（api/reviews.go の validateReview）と同じ制約を張る。
// server-only を付けないこと（クライアントコンポーネントから import するため）

// 評価は総合1軸（1〜5）のみ。技術力・コミュニケーションのような複数軸にすると
// 入力の手間が増えて提出率が下がる。両者提出が公開条件である以上、
// 提出率は機能が成立するかどうかを決める
export const RATING_VALUES = [1, 2, 3, 4, 5] as const

export const reviewSchema = z.object({
  // ラジオボタンの値は文字列で届くため coerce で数値にする
  rating: z.coerce
    .number()
    .int()
    .min(1, '評価を選択してください')
    .max(5, '評価は1〜5で選択してください'),
  comment: z
    .string()
    .trim()
    .min(1, 'コメントは必須です')
    .max(2000, 'コメントは2000文字以内にしてください'),
})

export type ReviewFormValues = z.input<typeof reviewSchema>
export type ReviewInputValues = z.output<typeof reviewSchema>

// 評価の言葉での説明。星の数だけでは基準が人によってぶれるため、
// 選択時に文言を添えて解釈を揃える
export const ratingLabels: Record<number, string> = {
  1: '期待を大きく下回った',
  2: '期待を下回った',
  3: '期待どおり',
  4: '期待を上回った',
  5: '期待を大きく上回った',
}
