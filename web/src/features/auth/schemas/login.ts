import type { z } from 'zod'

import { AuthLoginBody } from '@repo/api-client/company/generated/zod'

// ログイン入力は共通モデル（LoginInput）のため company / talent で形が同一。
// company 側の生成物を代表として使う。フォームと API の形も一致するため無加工
//（ADR-0007 案1: 一致するフィールドは生成スキーマをそのまま引き継ぐ、の最も単純なケース）
export const loginFormSchema = AuthLoginBody

export type LoginFormInput = z.input<typeof loginFormSchema>
export type LoginFormOutput = z.output<typeof loginFormSchema>
