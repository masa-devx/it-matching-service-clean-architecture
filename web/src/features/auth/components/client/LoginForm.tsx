'use client'

import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import Link from 'next/link'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

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

      <div>
        <label htmlFor="email" className="block font-medium">
          メールアドレス
        </label>
        <input
          id="email"
          type="email"
          {...register('email')}
          className="w-full rounded border p-2"
        />
        {errors.email && (
          <p role="alert" className="text-sm text-red-600">
            {errors.email.message}
          </p>
        )}
      </div>

      <div>
        <label htmlFor="password" className="block font-medium">
          パスワード
        </label>
        <input
          id="password"
          type="password"
          {...register('password')}
          className="w-full rounded border p-2"
        />
        {errors.password && (
          <p role="alert" className="text-sm text-red-600">
            {errors.password.message}
          </p>
        )}
      </div>

      <button
        type="submit"
        disabled={isSubmitting}
        className="rounded bg-blue-600 px-4 py-2 text-white disabled:opacity-50"
      >
        {isSubmitting ? 'ログイン中…' : 'ログイン'}
      </button>

      {serverError && (
        <p role="alert" className="font-medium text-red-600">
          {serverError}
        </p>
      )}

      <p className="text-sm">
        アカウントをお持ちでない方は{' '}
        <Link href={`/${role}/signup`} className="text-blue-600 underline">
          サインアップ
        </Link>
      </p>
    </form>
  )
}
