import { redirect } from 'next/navigation'

import { AppHeader } from '@/components/AppHeader'
import { LogoutButton } from '@/features/auth/components/client/LogoutButton'
import { meTalent } from '@/external/handler/auth'

// (talent) は talent ロール必須（company 側と対称）。
// アプリシェル（ヘッダー）の認証情報はここで解決し、判断済みの結果だけを渡す
export default async function TalentLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const me = await meTalent()
  if (!me) {
    redirect('/talent/login')
  }

  return (
    <div className="flex min-h-full flex-col">
      <AppHeader role="talent" displayName={me.display_name}>
        <LogoutButton role="talent" />
      </AppHeader>
      <main className="mx-auto w-full max-w-5xl flex-1 p-6">{children}</main>
    </div>
  )
}
