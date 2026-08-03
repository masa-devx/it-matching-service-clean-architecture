import Link from 'next/link'
import { LoginForm } from '@/components/LoginForm'

export const metadata = { title: 'ログイン | Tsunagu Works' }

export default function LoginPage() {
  return (
    <>
      <LoginForm />
      <p className="text-sm text-muted-foreground">
        アカウントをお持ちでない方は{' '}
        <Link href="/signup" className="text-primary underline">
          新規登録
        </Link>
      </p>
    </>
  )
}
