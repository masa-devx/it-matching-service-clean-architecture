'use server'

import { redirect } from 'next/navigation'

import { loginCompany, signupCompany } from '@/external/handler/auth'

import type { CompanySignupFormOutput } from '../schemas/companySignup'
import type { LoginFormOutput } from '../schemas/login'

// actions は「薄いラッパー」: handler を呼び、成功したら遷移するだけ。
// 通信・Cookie の詳細は external 層の責務（境界ルール: external/client は直接呼ばない）

export async function signupCompanyAction(
  data: CompanySignupFormOutput,
): Promise<{ error: string }> {
  const result = await signupCompany(data)
  if (!result.ok) {
    return { error: result.error }
  }
  redirect('/company/dashboard')
}

export async function loginCompanyAction(
  data: LoginFormOutput,
): Promise<{ error: string }> {
  const result = await loginCompany(data)
  if (!result.ok) {
    return { error: result.error }
  }
  redirect('/company/dashboard')
}
