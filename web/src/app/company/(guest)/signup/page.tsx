import Link from 'next/link'
import { SignupForm } from '@/components/SignupForm'

export const metadata = { title: '企業の新規登録 | Tsunagu Works' }

export default function CompanySignupPage() {
  return (
    <>
      <SignupForm role="company" />
      <div className="flex flex-col items-center gap-2 text-sm text-muted-foreground">
        <p>
          アカウントをお持ちの方は{' '}
          <Link href="/company/login" className="text-primary underline">
            ログイン
          </Link>
        </p>
        <p>
          <Link href="/talent/signup" className="underline">
            人材として登録する方はこちら
          </Link>
        </p>
      </div>
    </>
  )
}
