import { z } from 'zod'

// Go 側（api/messages.go の validateMessage）と同じ制約を張る。
// server-only を付けないこと（クライアントコンポーネントから import するため）
export const messageSchema = z.object({
  body: z
    .string()
    .trim()
    .min(1, 'メッセージを入力してください')
    .max(2000, 'メッセージは2000文字以内にしてください'),
})

export type MessageFormValues = z.infer<typeof messageSchema>
