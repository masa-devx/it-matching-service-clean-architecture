'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { useRouter } from 'next/navigation'
import {
  projectFormSchema,
  type ProjectFormValues,
  type ProjectInput,
} from '@/lib/projectSchema'
import { createProject } from '@/lib/projectClient'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

const emptyProject: ProjectFormValues = {
  title: '',
  description: '',
  required_skills: '',
  hourly_rate_min: 0,
  hourly_rate_max: 0,
  hours_per_week: 0,
  remote_ok: false,
  status: 'published',
}

export function ProjectForm() {
  const router = useRouter()
  const [serverError, setServerError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ProjectFormValues, unknown, ProjectInput>({
    resolver: standardSchemaResolver(projectFormSchema),
    defaultValues: emptyProject,
  })

  // 引数は Zod の transform 適用後（required_skills が string[] になっている）
  async function onSubmit(values: ProjectInput) {
    setServerError(null)

    const result = await createProject(values)
    if (!result.ok) {
      setServerError(result.error)
      return
    }

    // 作成できたら管理一覧へ。refresh で RSC を再実行し、追加した案件を反映する
    router.push('/company/projects')
    router.refresh()
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="flex w-full max-w-2xl flex-col gap-6"
      noValidate
    >
      <div className="flex flex-col gap-2">
        <Label htmlFor="title">
          案件タイトル <span className="text-destructive">*</span>
        </Label>
        <Input
          id="title"
          className="h-11"
          placeholder="例: Go APIの開発支援"
          {...register('title')}
        />
        {errors.title && (
          <p role="alert" className="text-sm text-destructive">
            {errors.title.message}
          </p>
        )}
      </div>

      <div className="flex flex-col gap-2">
        <Label htmlFor="description">案件内容</Label>
        <Textarea
          id="description"
          rows={6}
          placeholder="業務内容・技術スタック・体制など"
          {...register('description')}
        />
        {errors.description && (
          <p role="alert" className="text-sm text-destructive">
            {errors.description.message}
          </p>
        )}
      </div>

      <div className="flex flex-col gap-2">
        <Label htmlFor="required_skills">必須スキル</Label>
        <Input
          id="required_skills"
          className="h-11"
          placeholder="Go, PostgreSQL, AWS"
          {...register('required_skills')}
        />
        <p className="text-xs text-muted-foreground">
          カンマ区切りで入力してください（最大30個）
        </p>
        {errors.required_skills && (
          <p role="alert" className="text-sm text-destructive">
            {errors.required_skills.message}
          </p>
        )}
      </div>

      <div className="grid gap-6 sm:grid-cols-3">
        <div className="flex flex-col gap-2">
          <Label htmlFor="hourly_rate_min">時給の下限（円）</Label>
          <Input
            id="hourly_rate_min"
            type="number"
            inputMode="numeric"
            className="h-11"
            {...register('hourly_rate_min')}
          />
          {errors.hourly_rate_min && (
            <p role="alert" className="text-sm text-destructive">
              {errors.hourly_rate_min.message}
            </p>
          )}
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="hourly_rate_max">時給の上限（円）</Label>
          <Input
            id="hourly_rate_max"
            type="number"
            inputMode="numeric"
            className="h-11"
            {...register('hourly_rate_max')}
          />
          {errors.hourly_rate_max && (
            <p role="alert" className="text-sm text-destructive">
              {errors.hourly_rate_max.message}
            </p>
          )}
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="hours_per_week">週の稼働時間</Label>
          <Input
            id="hours_per_week"
            type="number"
            inputMode="numeric"
            className="h-11"
            {...register('hours_per_week')}
          />
          {errors.hours_per_week && (
            <p role="alert" className="text-sm text-destructive">
              {errors.hours_per_week.message}
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
        フルリモート可
      </label>

      <div className="flex flex-col gap-2">
        <Label htmlFor="status">公開設定</Label>
        <select
          id="status"
          className="h-11 rounded-lg border border-input bg-transparent px-2.5 text-base md:text-sm"
          {...register('status')}
        >
          <option value="published">すぐに公開する</option>
          <option value="draft">下書きとして保存する</option>
        </select>
      </div>

      {serverError && (
        <p role="alert" className="text-sm text-destructive">
          {serverError}
        </p>
      )}

      <Button type="submit" disabled={isSubmitting} className="h-11 self-start">
        {isSubmitting ? '掲載中…' : '案件を掲載する'}
      </Button>
    </form>
  )
}
