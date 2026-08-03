'use client'

import { useRouter } from 'next/navigation'
import { useState } from 'react'
import { signup } from '@/lib/authClient'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export function SignupForm() {
  const router = useRouter()
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)

    const formData = new FormData(e.currentTarget)
    const result = await signup({
      email: String(formData.get('email') ?? ''),
      password: String(formData.get('password') ?? ''),
      role: formData.get('role') === 'company' ? 'company' : 'talent',
    })

    if (!result.ok) {
      setError(result.error)
      setSubmitting(false)
      return
    }
    // 成功時は submitting を戻さない（遷移完了までボタンを無効のまま保つ）
    router.push('/')
  }

  return (
    <Card className="w-full max-w-sm">
      <CardHeader>
        <CardTitle>新規登録</CardTitle>
        <CardDescription>
          メールアドレスとパスワードで登録できます
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={handleSubmit}
          className="flex flex-col gap-5"
          noValidate={false}
        >
          <fieldset className="flex flex-col gap-2">
            <legend className="mb-2 text-sm font-medium">登録する立場</legend>
            <div className="flex gap-4">
              <label className="flex h-11 flex-1 cursor-pointer items-center justify-center gap-2 rounded-md border text-sm has-[:checked]:border-primary has-[:checked]:bg-primary/5 has-[:checked]:text-primary">
                <input
                  type="radio"
                  name="role"
                  value="talent"
                  defaultChecked
                  className="accent-primary"
                />
                人材（受注側）
              </label>
              <label className="flex h-11 flex-1 cursor-pointer items-center justify-center gap-2 rounded-md border text-sm has-[:checked]:border-primary has-[:checked]:bg-primary/5 has-[:checked]:text-primary">
                <input
                  type="radio"
                  name="role"
                  value="company"
                  className="accent-primary"
                />
                企業（発注側）
              </label>
            </div>
          </fieldset>

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

          <Button type="submit" disabled={submitting} className="h-11">
            {submitting ? '登録中…' : '登録する'}
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
