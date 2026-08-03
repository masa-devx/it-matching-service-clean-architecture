'use client'

import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { logout } from '@/lib/authClient'
import { Button } from '@/components/ui/button'

export function LogoutButton() {
  const router = useRouter()
  const [submitting, setSubmitting] = useState(false)

  async function handleLogout() {
    setSubmitting(true)
    await logout()
    // refresh でサーバーコンポーネントを再描画し、ログイン状態の表示を最新化する
    router.push('/login')
    router.refresh()
  }

  return (
    <Button
      variant="outline"
      onClick={handleLogout}
      disabled={submitting}
      className="h-11"
    >
      {submitting ? 'ログアウト中…' : 'ログアウト'}
    </Button>
  )
}
