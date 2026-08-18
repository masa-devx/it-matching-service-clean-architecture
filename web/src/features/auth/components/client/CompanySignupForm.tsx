'use client'

import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import Link from 'next/link'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

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
        <Label htmlFor="name">会社名</Label>
        <Input
          id="name"
          aria-invalid={errors.name ? true : undefined}
          {...register('name')}
        />
        {errors.name && (
          <p role="alert" className="text-sm text-destructive">
            {errors.name.message}
          </p>
        )}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="location">所在地（任意）</Label>
        <Input id="location" {...register('location')} />
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="description">事業内容（任意）</Label>
        <Textarea id="description" {...register('description')} />
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
        <Link href="/company/login" className="text-primary underline">
          ログイン
        </Link>
      </p>
    </form>
  )
}
