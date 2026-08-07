import { z } from 'zod'

// Go 側（api/work_reports.go の validateWorkReport）と同じ制約を張る。
// server-only を付けないこと（クライアントコンポーネントから import するため）

export const workReportSchema = z.object({
  hours: z.coerce
    .number()
    .int('整数で入力してください')
    // 1週間は168時間。それを超える申告は入力ミス
    .min(0, '稼働時間は0〜168の範囲で入力してください')
    .max(168, '稼働時間は0〜168の範囲で入力してください'),
  summary: z
    .string()
    .trim()
    .min(1, '作業内容は必須です')
    .max(2000, '作業内容は2000文字以内にしてください'),
  // 再提出では週を変更できない（週を変えるなら別の報告になる）が、
  // フォームの値としては保持する。入力欄は出さず、送信時も使わない
  // （api 側も PUT では週を受け取らない）
  week_start: z.string().min(1, '対象の週を選択してください'),
})

export type WorkReportFormValues = z.input<typeof workReportSchema>
export type WorkReportInputValues = z.output<typeof workReportSchema>

// 直近の週（月曜）を新しい順に返す。
//
// 稼働報告は「先週の分を出す」のが大半なので、選択肢を絞ったほうが速く選べる。
// ここで計算した値がずれても、保存される週は API 側が date_trunc('week', ...) で
// 丸めるため実害はない（クライアントの週計算を信用しない設計）
export function recentWeekStarts(count = 8, today = new Date()): string[] {
  const monday = new Date(today)
  // getDay() は日曜が0。月曜を週の開始とするため、日曜だけ6日戻す
  const offset = (monday.getDay() + 6) % 7
  monday.setDate(monday.getDate() - offset)

  const weeks: string[] = []
  for (let i = 0; i < count; i++) {
    weeks.push(formatDate(monday))
    monday.setDate(monday.getDate() - 7)
  }
  return weeks
}

// 「8月3日(月)〜8月9日(日)」のように週の範囲を表示する。
// 開始日だけだと「その週」だと分かりにくいため、終わりまで見せる
export function formatWeekRange(weekStart: string): string {
  const start = new Date(`${weekStart}T00:00:00`)
  const end = new Date(start)
  end.setDate(end.getDate() + 6)

  const format = (d: Date) =>
    d.toLocaleDateString('ja-JP', { month: 'long', day: 'numeric' })
  return `${format(start)}〜${format(end)}`
}

// ローカル時刻のまま YYYY-MM-DD にする。
// toISOString() だと UTC に変換されるため、日本時間の早朝に前日の日付になってしまう
function formatDate(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
