'use server'

import { redirect } from 'next/navigation'

import { loginTalent, signupTalent } from '@/external/handler/auth'

import type { LoginFormOutput } from '../schemas/login'
import type { TalentSignupFormOutput } from '../schemas/talentSignup'

export async function signupTalentAction(
  data: TalentSignupFormOutput,
): Promise<{ error: string }> {
  const result = await signupTalent(data)
  if (!result.ok) {
    return { error: result.error }
  }
  redirect('/talent/dashboard')
}

export async function loginTalentAction(
  data: LoginFormOutput,
): Promise<{ error: string }> {
  const result = await loginTalent(data)
  if (!result.ok) {
    return { error: result.error }
  }
  redirect('/talent/dashboard')
}
