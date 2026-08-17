'use client'

import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import Link from 'next/link'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { signupTalentAction } from '../../actions/talent'
import {
  talentSignupFormSchema,
  type TalentSignupFormInput,
  type TalentSignupFormOutput,
} from '../../schemas/talentSignup'

export function TalentSignupForm() {
  const [serverError, setServerError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<TalentSignupFormInput, unknown, TalentSignupFormOutput>({
    resolver: standardSchemaResolver(talentSignupFormSchema),
    defaultValues: {
      email: '',
      password: '',
      display_name: '',
      skills: '',
      bio: '',
    },
  })

  const onSubmit = async (data: TalentSignupFormOutput) => {
    setServerError(null)
    const result = await signupTalentAction(data)
    if (result?.error) {
      setServerError(result.error)
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="max-w-md space-y-4">
      <h1 className="text-2xl font-bold">人材サインアップ</h1>

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
        <label htmlFor="display_name" className="block font-medium">
          表示名
        </label>
        <input
          id="display_name"
          {...register('display_name')}
          className="w-full rounded border p-2"
        />
        {errors.display_name && (
          <p role="alert" className="text-sm text-red-600">
            {errors.display_name.message}
          </p>
        )}
      </div>

      <div>
        <label htmlFor="skills" className="block font-medium">
          スキル（カンマ区切り）
        </label>
        <input
          id="skills"
          placeholder="Go, TypeScript"
          {...register('skills')}
          className="w-full rounded border p-2"
        />
      </div>

      <div>
        <label htmlFor="bio" className="block font-medium">
          自己紹介（任意）
        </label>
        <textarea
          id="bio"
          {...register('bio')}
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
        <Link href="/talent/login" className="text-blue-600 underline">
          ログイン
        </Link>
      </p>
    </form>
  )
}
