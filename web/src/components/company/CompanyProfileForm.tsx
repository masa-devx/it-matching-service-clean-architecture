'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { useRouter } from 'next/navigation'
import {
  companyProfileSchema,
  type CompanyProfileInput,
} from '@/lib/profileSchema'
import { saveProfile } from '@/lib/profileClient'
import type { CompanyProfile } from '@/lib/profile'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

const emptyProfile: CompanyProfileInput = {
  name: '',
  description: '',
  industry: '',
  size: '',
}

export function CompanyProfileForm({
  profile,
}: {
  profile: CompanyProfile | null
}) {
  const router = useRouter()
  const [saved, setSaved] = useState(false)
  const [serverError, setServerError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<CompanyProfileInput>({
    resolver: standardSchemaResolver(companyProfileSchema),
    // 未作成なら空、作成済みなら現在値を初期表示する
    defaultValues: profile ?? emptyProfile,
  })

  // handleSubmit が検証を通した後にだけ呼ばれる（引数は検証済みの値）
  async function onSubmit(values: CompanyProfileInput) {
    setSaved(false)
    setServerError(null)

    const result = await saveProfile(values)
    if (!result.ok) {
      setServerError(result.error)
      return
    }

    setSaved(true)
    // RSC（ページ）を再取得して表示を最新化する
    router.refresh()
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="flex w-full max-w-2xl flex-col gap-6"
      noValidate
    >
      <div className="flex flex-col gap-2">
        <Label htmlFor="name">
          会社名 <span className="text-destructive">*</span>
        </Label>
        <Input id="name" className="h-11" {...register('name')} />
        {errors.name && (
          <p role="alert" className="text-sm text-destructive">
            {errors.name.message}
          </p>
        )}
      </div>

      <div className="flex flex-col gap-2">
        <Label htmlFor="description">会社説明</Label>
        <Textarea id="description" rows={5} {...register('description')} />
        {errors.description && (
          <p role="alert" className="text-sm text-destructive">
            {errors.description.message}
          </p>
        )}
      </div>

      <div className="flex flex-col gap-2">
        <Label htmlFor="industry">業種</Label>
        <Input
          id="industry"
          className="h-11"
          placeholder="例: SaaS / 受託開発"
          {...register('industry')}
        />
        {errors.industry && (
          <p role="alert" className="text-sm text-destructive">
            {errors.industry.message}
          </p>
        )}
      </div>

      <div className="flex flex-col gap-2">
        <Label htmlFor="size">従業員規模</Label>
        <Input
          id="size"
          className="h-11"
          placeholder="例: 11-50"
          {...register('size')}
        />
        {errors.size && (
          <p role="alert" className="text-sm text-destructive">
            {errors.size.message}
          </p>
        )}
      </div>

      {serverError && (
        <p role="alert" className="text-sm text-destructive">
          {serverError}
        </p>
      )}
      {saved && (
        <p role="status" className="text-sm text-primary">
          保存しました
        </p>
      )}

      <Button type="submit" disabled={isSubmitting} className="h-11 self-start">
        {isSubmitting ? '保存中…' : '保存する'}
      </Button>
    </form>
  )
}
