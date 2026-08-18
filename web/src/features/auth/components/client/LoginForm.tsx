'use client'

import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import Link from 'next/link'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

import { loginCompanyAction } from '../../actions/company'
import { loginTalentAction } from '../../actions/talent'
import {
  loginFormSchema,
  type LoginFormInput,
  type LoginFormOutput,
} from '../../schemas/login'

type Props = {
  role: 'company' | 'talent'
}

const roleLabel = { company: '企業', talent: '人材' } as const

export function LoginForm({ role }: Props) {
  const [serverError, setServerError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormInput, unknown, LoginFormOutput>({
    resolver: standardSchemaResolver(loginFormSchema),
    defaultValues: { email: '', password: '' },
  })

  const onSubmit = async (data: LoginFormOutput) => {
    setServerError(null)
    const action = role === 'company' ? loginCompanyAction : loginTalentAction
    // 成功時は action 内の redirect で遷移するため、戻り値があるのは失敗時だけ
    const result = await action(data)
    if (result?.error) {
      setServerError(result.error)
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="max-w-md space-y-4">
      <h1 className="text-2xl font-bold">{roleLabel[role]}ログイン</h1>

      <div className="space-y-1.5">
        <Label htmlFor="email">メールアドレス</Label>
        <Input
          id="email"
          type="email"
          aria-invalid={errors.email ? true : undefined}
          {...register('email')}
        />
        {errors.email && (
          <p role="alert" className="text-sm text-destructive">
            {errors.email.message}
          </p>
        )}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="password">パスワード</Label>
        <Input
          id="password"
          type="password"
          aria-invalid={errors.password ? true : undefined}
          {...register('password')}
        />
        {errors.password && (
          <p role="alert" className="text-sm text-destructive">
            {errors.password.message}
          </p>
        )}
      </div>

      <Button type="submit" disabled={isSubmitting}>
        {isSubmitting ? 'ログイン中…' : 'ログイン'}
      </Button>

      {serverError && (
        <p role="alert" className="font-medium text-destructive">
          {serverError}
        </p>
      )}

      <p className="text-sm">
        アカウントをお持ちでない方は{' '}
        <Link href={`/${role}/signup`} className="text-primary underline">
          サインアップ
        </Link>
      </p>
    </form>
  )
}
