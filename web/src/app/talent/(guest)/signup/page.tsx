import Link from 'next/link'
import { SignupForm } from '@/components/SignupForm'

export const metadata = { title: '人材の新規登録 | Tsunagu Works' }

export default function TalentSignupPage() {
  return (
    <>
      <SignupForm role="talent" />
      <div className="flex flex-col items-center gap-2 text-sm text-muted-foreground">
        <p>
          アカウントをお持ちの方は{' '}
          <Link href="/talent/login" className="text-primary underline">
            ログイン
          </Link>
        </p>
        <p>
          <Link href="/company/signup" className="underline">
            企業として登録する方はこちら
          </Link>
        </p>
      </div>
    </>
  )
}
