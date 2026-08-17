'use client'

import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import Link from 'next/link'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { signupCompanyAction } from '../../actions/company'
import {
  companySignupFormSchema,
  type CompanySignupFormInput,
  type CompanySignupFormOutput,
} from '../../schemas/companySignup'

export function CompanySignupForm() {
  const [serverError, setServerError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<CompanySignupFormInput, unknown, CompanySignupFormOutput>({
    resolver: standardSchemaResolver(companySignupFormSchema),
    defaultValues: {
      email: '',
      password: '',
      name: '',
      location: '',
      description: '',
    },
  })

  const onSubmit = async (data: CompanySignupFormOutput) => {
    setServerError(null)
    const result = await signupCompanyAction(data)
    if (result?.error) {
      setServerError(result.error)
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="max-w-md space-y-4">
      <h1 className="text-2xl font-bold">企業サインアップ</h1>

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
          パスワード（8文字以上）
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

      <div>
        <label htmlFor="name" className="block font-medium">
          会社名
        </label>
        <input
          id="name"
          {...register('name')}
          className="w-full rounded border p-2"
        />
        {errors.name && (
          <p role="alert" className="text-sm text-red-600">
            {errors.name.message}
          </p>
        )}
      </div>

      <div>
        <label htmlFor="location" className="block font-medium">
          所在地（任意）
        </label>
        <input
          id="location"
          {...register('location')}
          className="w-full rounded border p-2"
        />
      </div>

      <div>
        <label htmlFor="description" className="block font-medium">
          事業内容（任意）
        </label>
        <textarea
          id="description"
          {...register('description')}
          className="w-full rounded border p-2"
        />
      </div>

      <button
        type="submit"
        disabled={isSubmitting}
        className="rounded bg-blue-600 px-4 py-2 text-white disabled:opacity-50"
      >
        {isSubmitting ? '登録中…' : 'サインアップ'}
      </button>

      {serverError && (
        <p role="alert" className="font-medium text-red-600">
          {serverError}
        </p>
      )}

      <p className="text-sm">
        既にアカウントをお持ちの方は{' '}
        <Link href="/company/login" className="text-blue-600 underline">
          ログイン
        </Link>
      </p>
    </form>
  )
}
