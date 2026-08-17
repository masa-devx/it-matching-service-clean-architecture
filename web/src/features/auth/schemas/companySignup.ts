import type { z } from 'zod'

import { AuthSignupBody } from '@repo/api-client/company/generated/zod'

// 企業サインアップ: フォームと API の形が一致するため無加工（location / description は
// optional の文字列で、未入力は空文字のまま送る＝「未設定の文字列は空文字」の方針）
export const companySignupFormSchema = AuthSignupBody

export type CompanySignupFormInput = z.input<typeof companySignupFormSchema>
export type CompanySignupFormOutput = z.output<typeof companySignupFormSchema>
