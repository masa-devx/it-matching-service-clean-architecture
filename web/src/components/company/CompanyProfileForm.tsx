'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import { useForm } from 'react-hook-form'
import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { useRouter } from 'next/navigation'
import {
  companyProfileSchema,
  type CompanyProfileInput,
} from '@/lib/profileSchema'
import { saveProfile } from '@/lib/profileClient'
import type { CompanyProfile } from '@/lib/profile'
import { SubmitButton } from '@/components/SubmitButton'
import { FormErrorSummary } from '@/components/FormErrorSummary'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RequiredMark } from '@/components/RequiredMark'
import { Textarea } from '@/components/ui/textarea'

const emptyProfile: CompanyProfileInput = {
  name: '',
  description: '',
  industry: '',
  size: '',
}

// エラーサマリで内部名ではなく画面上のラベルを見せるための対応表
const fieldLabels = {
  name: '会社名',
  description: '会社説明',
  industry: '業種',
  size: '従業員規模',
}

export function CompanyProfileForm({
  profile,
}: {
  profile: CompanyProfile | null
}) {
  const router = useRouter()
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
    setServerError(null)

    const result = await saveProfile(values)
    if (!result.ok) {
      setServerError(result.error)
      toast.error(result.error)
      return
    }

    // 保存の成否はトーストで通知する（本文中のテキストは見落としやすい）
    toast.success('プロフィールを保存しました')
    // RSC（ページ）を再取得して表示を最新化する
    router.refresh()
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="flex w-full max-w-2xl flex-col gap-6"
      noValidate
    >
      <FormErrorSummary errors={errors} labels={fieldLabels} />

      <div className="flex flex-col gap-2">
        <Label htmlFor="name">
          会社名 <RequiredMark />
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

      <SubmitButton isSubmitting={isSubmitting} submittingLabel="保存中…">
        保存する
      </SubmitButton>
    </form>
  )
}
