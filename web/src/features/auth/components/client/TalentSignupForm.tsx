'use client'

import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import Link from 'next/link'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

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
        <Label htmlFor="password">パスワード（8文字以上）</Label>
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

      <div className="space-y-1.5">
        <Label htmlFor="display_name">表示名</Label>
        <Input
          id="display_name"
          aria-invalid={errors.display_name ? true : undefined}
          {...register('display_name')}
        />
        {errors.display_name && (
          <p role="alert" className="text-sm text-destructive">
            {errors.display_name.message}
          </p>
        )}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="skills">スキル（カンマ区切り）</Label>
        <Input
          id="skills"
          placeholder="Go, TypeScript"
          {...register('skills')}
        />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="bio">自己紹介（任意）</Label>
        <Textarea id="bio" {...register('bio')} />
      </div>

      <Button type="submit" disabled={isSubmitting}>
        {isSubmitting ? '登録中…' : 'サインアップ'}
      </Button>

      {serverError && (
        <p role="alert" className="font-medium text-destructive">
          {serverError}
        </p>
      )}

      <p className="text-sm">
        既にアカウントをお持ちの方は{' '}
        <Link href="/talent/login" className="text-primary underline">
          ログイン
        </Link>
      </p>
    </form>
  )
}
