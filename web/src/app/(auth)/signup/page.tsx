import Link from 'next/link'
import { SignupForm } from '@/components/SignupForm'

export const metadata = { title: '新規登録 | Tsunagu Works' }

export default function SignupPage() {
  return (
    <>
      <SignupForm />
      <p className="text-sm text-muted-foreground">
        アカウントをお持ちの方は{' '}
        <Link href="/login" className="text-primary underline">
          ログイン
        </Link>
      </p>
    </>
  )
}
