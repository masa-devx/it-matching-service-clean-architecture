import { z } from 'zod'

import { AuthSignupBody } from '@repo/api-client/talent/generated/zod'

// 人材サインアップ: skills だけフォーム（カンマ区切り文字列）と API（配列）で形が違うため
// .omit().extend() で差し替える（ADR-0007 案1・project フォームと同じパターン）
export const talentSignupFormSchema = AuthSignupBody.omit({
  skills: true,
}).extend({
  skills: z.string().transform((v) =>
    v
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean),
  ),
})

export type TalentSignupFormInput = z.input<typeof talentSignupFormSchema>
export type TalentSignupFormOutput = z.output<typeof talentSignupFormSchema>
