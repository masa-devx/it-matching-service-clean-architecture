import { z } from 'zod'
import { ProjectsCreateBody } from '@repo/api-client/company/generated/zod'

// 生成Zodを加工してフォーム用スキーマを作る（ADR-0007 案1）。
// API と形が一致するフィールドは生成スキーマをそのまま引き継ぎ（制約も含めて二重管理しない）、
// フォームと API で本質的に形が違うものだけを差し替える
export const projectFormSchema = ProjectsCreateBody.omit({
  // フォームではカンマ区切りの1テキスト入力にするため、配列定義を一旦外す
  required_skills: true,
}).extend({
  // "Go, PostgreSQL" → ["Go", "PostgreSQL"]（空要素は除去）
  required_skills: z.string().transform((v) =>
    v
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean),
  ),
})

// z.input = フォームが保持する形（required_skills は string）
// z.output = parse 後 = API に送る形（required_skills は string[]）
export type ProjectFormInput = z.input<typeof projectFormSchema>
export type ProjectFormOutput = z.output<typeof projectFormSchema>
