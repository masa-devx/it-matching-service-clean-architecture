import { redirect } from 'next/navigation'

import { meTalent } from '@/external/handler/auth'

// (talent) は talent ロール必須（company 側と対称）
export default async function TalentLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const me = await meTalent()
  if (!me) {
    redirect('/talent/login')
  }

  return <main className="mx-auto max-w-2xl p-8">{children}</main>
}
