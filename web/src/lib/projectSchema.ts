import { z } from 'zod'

// 案件掲載フォームの検証ルール。Go 側（api/projects.go の validateProject）と
// 同じ制約・同じ文言を張る（フロント=即時フィードバック / サーバー=最後の防衛線）

export const projectStatuses = ['draft', 'published'] as const

export const projectFormSchema = z
  .object({
    title: z
      .string()
      .trim()
      .min(1, '案件タイトルは必須です')
      .max(100, '案件タイトルは100文字以内にしてください'),
    description: z.string().max(5000, '案件内容は5000文字以内にしてください'),
    // スキルはカンマ区切りの1入力で受け、送信時に配列へ変換する
    required_skills: z
      .string()
      .transform(splitSkills)
      .refine((v) => v.length <= 30, '必須スキルは30個以内にしてください')
      .refine(
        (v) => v.every((s) => s.length <= 50),
        '各スキルは50文字以内にしてください',
      ),
    hourly_rate_min: z.coerce
      .number()
      .int('整数で入力してください')
      .min(0, '単価は0〜1000000の範囲で入力してください')
      .max(1000000, '単価は0〜1000000の範囲で入力してください'),
    hourly_rate_max: z.coerce
      .number()
      .int('整数で入力してください')
      .min(0, '単価は0〜1000000の範囲で入力してください')
      .max(1000000, '単価は0〜1000000の範囲で入力してください'),
    hours_per_week: z.coerce
      .number()
      .int('整数で入力してください')
      .min(0, '週の稼働時間は0〜168の範囲で入力してください')
      .max(168, '週の稼働時間は0〜168の範囲で入力してください'),
    remote_ok: z.boolean(),
    status: z.enum(projectStatuses),
  })
  // 複数フィールドにまたがる検証は refine で行い、エラーを該当フィールドに紐づける
  .refine((v) => v.hourly_rate_min <= v.hourly_rate_max, {
    message: '単価の下限は上限以下にしてください',
    path: ['hourly_rate_max'],
  })

export type ProjectFormValues = z.input<typeof projectFormSchema>
export type ProjectInput = z.output<typeof projectFormSchema>

// Go 側 normalizeSkills と同じ規則（trim・空要素除去・重複排除）
function splitSkills(value: string): string[] {
  const seen = new Set<string>()
  return value
    .split(',')
    .map((s) => s.trim())
    .filter((s) => s !== '' && !seen.has(s) && seen.add(s) !== undefined)
}
