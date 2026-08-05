'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import { useForm } from 'react-hook-form'
import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { useRouter } from 'next/navigation'
import {
  talentProfileSchema,
  joinSkills,
  type TalentProfileForm as TalentProfileFormValues,
  type TalentProfileInput,
} from '@/lib/profileSchema'
import { saveProfile } from '@/lib/profileClient'
import type { TalentProfile } from '@/lib/profile'
import { SubmitButton } from '@/components/SubmitButton'
import { FormErrorSummary } from '@/components/FormErrorSummary'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

const emptyProfile: TalentProfileFormValues = {
  bio: '',
  skills: '',
  years_of_exp: 0,
  available_hours_per_week: 0,
  desired_hourly_rate: 0,
  remote_ok: false,
}

// APIの形（skills が配列）をフォームの形（skills がカンマ区切り文字列）に変換する
function toFormValues(profile: TalentProfile): TalentProfileFormValues {
  return {
    bio: profile.bio,
    skills: joinSkills(profile.skills),
    years_of_exp: profile.years_of_exp,
    available_hours_per_week: profile.available_hours_per_week,
    desired_hourly_rate: profile.desired_hourly_rate,
    remote_ok: profile.remote_ok,
  }
}

// エラーサマリで内部名ではなく画面上のラベルを見せるための対応表
const fieldLabels = {
  bio: '自己紹介',
  skills: 'スキル',
  years_of_exp: '経験年数',
  available_hours_per_week: '稼働可能時間',
  desired_hourly_rate: '希望時給',
}

export function TalentProfileForm({
  profile,
}: {
  profile: TalentProfile | null
}) {
  const router = useRouter()
  const [serverError, setServerError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<TalentProfileFormValues, unknown, TalentProfileInput>({
    resolver: standardSchemaResolver(talentProfileSchema),
    defaultValues: profile ? toFormValues(profile) : emptyProfile,
  })

  // 引数は Zod の transform 適用後（skills が string[] になっている）
  async function onSubmit(values: TalentProfileInput) {
    setServerError(null)

    const result = await saveProfile(values)
    if (!result.ok) {
      setServerError(result.error)
      toast.error(result.error)
      return
    }

    // 保存の成否はトーストで通知する（本文中のテキストは見落としやすい）
    toast.success('プロフィールを保存しました')
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
        <Label htmlFor="bio">自己紹介</Label>
        <Textarea
          id="bio"
          rows={5}
          placeholder="経歴・得意分野・実績など"
          {...register('bio')}
        />
        {errors.bio && (
          <p role="alert" className="text-sm text-destructive">
            {errors.bio.message}
          </p>
        )}
      </div>

      <div className="flex flex-col gap-2">
        <Label htmlFor="skills">スキル</Label>
        <Input
          id="skills"
          className="h-11"
          placeholder="Go, React, AWS"
          {...register('skills')}
        />
        <p className="text-xs text-muted-foreground">
          カンマ区切りで入力してください（最大30個）
        </p>
        {errors.skills && (
          <p role="alert" className="text-sm text-destructive">
            {errors.skills.message}
          </p>
        )}
      </div>

      <div className="grid gap-6 sm:grid-cols-3">
        <div className="flex flex-col gap-2">
          <Label htmlFor="years_of_exp">経験年数</Label>
          <Input
            id="years_of_exp"
            type="number"
            inputMode="numeric"
            className="h-11"
            {...register('years_of_exp')}
          />
          {errors.years_of_exp && (
            <p role="alert" className="text-sm text-destructive">
              {errors.years_of_exp.message}
            </p>
          )}
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="available_hours_per_week">稼働可能（時間/週）</Label>
          <Input
            id="available_hours_per_week"
            type="number"
            inputMode="numeric"
            className="h-11"
            {...register('available_hours_per_week')}
          />
          {errors.available_hours_per_week && (
            <p role="alert" className="text-sm text-destructive">
              {errors.available_hours_per_week.message}
            </p>
          )}
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="desired_hourly_rate">希望時給（円）</Label>
          <Input
            id="desired_hourly_rate"
            type="number"
            inputMode="numeric"
            className="h-11"
            {...register('desired_hourly_rate')}
          />
          {errors.desired_hourly_rate && (
            <p role="alert" className="text-sm text-destructive">
              {errors.desired_hourly_rate.message}
            </p>
          )}
        </div>
      </div>

      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          className="size-4 accent-primary"
          {...register('remote_ok')}
        />
        フルリモート勤務が可能
      </label>

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
