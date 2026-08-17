'use server'

import { redirect } from 'next/navigation'

import { logout } from '@/external/handler/auth'

export async function logoutAction(role: 'company' | 'talent'): Promise<void> {
  await logout()
  redirect(`/${role}/login`)
}
