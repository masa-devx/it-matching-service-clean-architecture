import { z } from 'zod'

// Go 側（api/applications.go の validateApplication）と同じ制約を張る。
// server-only を付けないこと（クライアントコンポーネントから import するため）

export const applicationSchema = z.object({
  message: z
    .string()
    .trim()
    .min(1, '志望動機は必須です')
    .max(2000, '志望動機は2000文字以内にしてください'),
})

export type ApplicationInput = z.infer<typeof applicationSchema>
