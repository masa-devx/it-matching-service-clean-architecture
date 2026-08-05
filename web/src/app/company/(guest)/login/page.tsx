import Link from 'next/link'
import { LoginForm } from '@/components/LoginForm'

export const metadata = { title: '企業ログイン | Tsunagu Works' }

export default function CompanyLoginPage() {
  return (
    <>
      <LoginForm role="company" />
      <div className="flex flex-col items-center gap-2 text-sm text-muted-foreground">
        <p>
          アカウントをお持ちでない方は{' '}
          <Link href="/company/signup" className="text-primary underline">
            新規登録
          </Link>
        </p>
        <p>
          <Link href="/talent/login" className="underline">
            人材の方はこちら
          </Link>
        </p>
      </div>
    </>
  )
}
