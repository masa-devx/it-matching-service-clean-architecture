import { z } from 'zod'

// フォームの検証ルール。Go 側（api/profile.go の validateXxxProfile）と同じ制約を張る。
// 二重バリデーション: フロントは「即時フィードバック」・サーバーは「最後の防衛線」。
// server-only を付けないこと（クライアントコンポーネントからも import するため）

export const companyProfileSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, '会社名は必須です')
    .max(100, '会社名は100文字以内にしてください'),
  description: z.string().max(2000, '会社説明は2000文字以内にしてください'),
  industry: z.string().max(50, '業種は50文字以内にしてください'),
  size: z.string().max(50, '従業員規模は50文字以内にしてください'),
})

export type CompanyProfileInput = z.infer<typeof companyProfileSchema>

// スキルはカンマ区切りの1入力で受け、送信時に配列へ変換する（タグUIは過剰）
export const talentProfileSchema = z.object({
  // 企業が応募者を識別する唯一の手がかりなので必須（連絡先は表示されないため）
  display_name: z
    .string()
    .trim()
    .min(1, '表示名は必須です')
    .max(50, '表示名は50文字以内にしてください'),
  bio: z.string().max(2000, '自己紹介は2000文字以内にしてください'),
  skills: z
    .string()
    .transform(splitSkills)
    .refine((v) => v.length <= 30, 'スキルは30個以内にしてください')
    .refine(
      (v) => v.every((s) => s.length <= 50),
      '各スキルは50文字以内にしてください',
    ),
  years_of_exp: z.coerce
    .number()
    .int('整数で入力してください')
    .min(0, '経験年数は0〜70の範囲で入力してください')
    .max(70, '経験年数は0〜70の範囲で入力してください'),
  available_hours_per_week: z.coerce
    .number()
    .int('整数で入力してください')
    // 1週間は168時間。それを超える稼働は入力ミス
    .min(0, '週の稼働可能時間は0〜168の範囲で入力してください')
    .max(168, '週の稼働可能時間は0〜168の範囲で入力してください'),
  desired_hourly_rate: z.coerce
    .number()
    .int('整数で入力してください')
    .min(0, '希望時給は0〜1000000の範囲で入力してください')
    .max(1000000, '希望時給は0〜1000000の範囲で入力してください'),
  remote_ok: z.boolean(),
})

// 変換前（フォームが扱う型）と変換後（APIへ送る型）
export type TalentProfileForm = z.input<typeof talentProfileSchema>
export type TalentProfileInput = z.output<typeof talentProfileSchema>

// 入力の揺れ（前後の空白・空要素・重複）を吸収する。Go 側の normalizeSkills と同じ規則
function splitSkills(value: string): string[] {
  const seen = new Set<string>()
  return value
    .split(',')
    .map((s) => s.trim())
    .filter((s) => s !== '' && !seen.has(s) && seen.add(s) !== undefined)
}

// APIから受け取った配列を、フォームの入力欄（カンマ区切り）へ戻す
export function joinSkills(skills: string[]): string {
  return skills.join(', ')
}
