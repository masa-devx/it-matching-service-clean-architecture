'use client'

import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { toast } from 'sonner'
import { signup } from '@/lib/authClient'
import { authCopyByRole } from '@/lib/nav'
import type { CurrentUser } from '@/lib/auth'
import { SubmitButton } from '@/components/SubmitButton'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export function SignupForm({ role }: { role: CurrentUser['role'] }) {
  const copy = authCopyByRole[role]

  const router = useRouter()
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)

    const formData = new FormData(e.currentTarget)
    // ロールはURL（画面）で決まるため、フォームでは選ばせない
    const result = await signup({
      email: String(formData.get('email') ?? ''),
      password: String(formData.get('password') ?? ''),
      role,
    })

    if (!result.ok) {
      setError(result.error)
      toast.error(result.error)
      setSubmitting(false)
      return
    }
    toast.success('登録が完了しました')
    // 成功時は submitting を戻さない（遷移完了までボタンを無効のまま保つ）
    router.push(result.redirectTo)
  }

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle>{copy.signupTitle}</CardTitle>
        <CardDescription>{copy.signupDescription}</CardDescription>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={handleSubmit}
          className="flex flex-col gap-5"
          noValidate={false}
        >
          <div className="flex flex-col gap-2">
            <Label htmlFor="email">メールアドレス</Label>
            <Input
              id="email"
              name="email"
              type="email"
              autoComplete="email"
              required
              placeholder="you@example.com"
              className="h-11"
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="password">パスワード</Label>
            <Input
              id="password"
              name="password"
              type="password"
              autoComplete="new-password"
              required
              minLength={8}
              placeholder="8文字以上"
              className="h-11"
            />
          </div>

          {error && (
            <p role="alert" className="text-sm text-destructive">
              {error}
            </p>
          )}

          <SubmitButton
            isSubmitting={submitting}
            submittingLabel="登録中…"
            className="h-11"
          >
            登録する
          </SubmitButton>
        </form>
      </CardContent>
    </Card>
  )
}
