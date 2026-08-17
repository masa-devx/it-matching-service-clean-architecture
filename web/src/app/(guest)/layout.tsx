import { redirect } from 'next/navigation'

import { currentRole } from '@/external/handler/auth'

// (guest) は未ログイン専用。ログイン済みならロール別ホームへ送り返す。
// 判定は Cookie の有無ではなく Go の me（currentRole 内）を単一の真実として使う
export default async function GuestLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const role = await currentRole()
  if (role === 'company') {
    redirect('/company/dashboard')
  }
  if (role === 'talent') {
    redirect('/talent/dashboard')
  }

  return <main className="mx-auto max-w-2xl p-8">{children}</main>
}
