import Link from 'next/link'
import { LoginForm } from '@/components/LoginForm'

export const metadata = { title: '人材ログイン | Tsunagu Works' }

export default function TalentLoginPage() {
  return (
    <>
      <LoginForm role="talent" />
      <div className="flex flex-col items-center gap-2 text-sm text-muted-foreground">
        <p>
          アカウントをお持ちでない方は{' '}
          <Link href="/talent/signup" className="text-primary underline">
            新規登録
          </Link>
        </p>
        <p>
          <Link href="/company/login" className="underline">
            企業の方はこちら
          </Link>
        </p>
      </div>
    </>
  )
}
