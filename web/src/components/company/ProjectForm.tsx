'use client'

import { useState } from 'react'
import { toast } from 'sonner'
import { Controller, useForm } from 'react-hook-form'
import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { useRouter } from 'next/navigation'
import {
  projectFormSchema,
  type ProjectFormValues,
  type ProjectInput,
} from '@/lib/projectSchema'
import { createProject, updateProject } from '@/lib/projectClient'
import { joinSkills } from '@/lib/projectSchema'
import type { MyProject } from '@/lib/companyProjects'
import { SubmitButton } from '@/components/SubmitButton'
import { FormErrorSummary } from '@/components/FormErrorSummary'
import { useUnsavedChangesWarning } from '@/hooks/useUnsavedChangesWarning'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { RequiredMark } from '@/components/RequiredMark'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

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

// APIの形（required_skills が配列）をフォームの形（カンマ区切り文字列）に変換する。
// status はフォームに出さないが、スキーマの型を満たすため現在値を入れておく
// （PUT では Go 側が status を無視するため、送っても掲載状態は変わらない）
function toFormValues(project: MyProject): ProjectFormValues {
  return {
    title: project.title,
    description: project.description,
    required_skills: joinSkills(project.required_skills),
    hourly_rate_min: project.hourly_rate_min,
    hourly_rate_max: project.hourly_rate_max,
    hours_per_week: project.hours_per_week,
    remote_ok: project.remote_ok,
    status: project.status === 'draft' ? 'draft' : 'published',
  }
}

// エラーサマリで内部名ではなく画面上のラベルを見せるための対応表
const fieldLabels = {
  title: '案件タイトル',
  description: '案件内容',
  required_skills: '必須スキル',
  hourly_rate_min: '時給の下限',
  hourly_rate_max: '時給の上限',
  hours_per_week: '週の稼働時間',
  status: '公開設定',
}

// project を渡すと編集モードになる。新規と編集で入力項目は同じだが、
// 掲載状態だけは扱いが違う（新規＝初期状態を選ぶ / 編集＝詳細画面の専用操作で変える）
export function ProjectForm({ project }: { project?: MyProject }) {
  const router = useRouter()
  const [serverError, setServerError] = useState<string | null>(null)
  const isEdit = project != null

  const {
    register,
    handleSubmit,
    control,
    formState: { errors, isSubmitting, isDirty },
  } = useForm<ProjectFormValues, unknown, ProjectInput>({
    resolver: standardSchemaResolver(projectFormSchema),
    defaultValues: project ? toFormValues(project) : emptyProject,
  })

  // 入力途中の離脱で内容が失われるのを防ぐ（掲載フォームは入力量が多い）
  useUnsavedChangesWarning(isDirty && !isSubmitting)

  // 引数は Zod の transform 適用後（required_skills が string[] になっている）
  async function onSubmit(values: ProjectInput) {
    setServerError(null)

    const result = isEdit
      ? await updateProject(project.id, values)
      : await createProject(values)
    if (!result.ok) {
      setServerError(result.error)
      toast.error(result.error)
      return
    }

    toast.success(isEdit ? '案件を更新しました' : '案件を掲載しました')

    // 保存できたら一覧（新規）または詳細（編集）へ。
    // refresh で RSC を再実行し、変更を反映する
    router.push(
      isEdit ? `/company/projects/${project.id}` : '/company/projects',
    )
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
        <Label htmlFor="title">
          案件タイトル <RequiredMark />
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

      {/* 公開設定は新規作成のときだけ。編集で出すと「文言を直しただけで
          意図せず公開される」事故が起きるため、掲載状態の変更は詳細画面の専用操作に分けている */}
      {!isEdit && (
        <div className="flex flex-col gap-2">
          <Label htmlFor="status">公開設定</Label>
          {/* Radix の Select は <select> ではないため register では繋がらない。
              Controller で RHF の管理下に置く（キーボード操作とa11yは部品側が担保） */}
          <Controller
            name="status"
            control={control}
            render={({ field }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id="status" className="h-11">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="published">すぐに公開する</SelectItem>
                  <SelectItem value="draft">下書きとして保存する</SelectItem>
                </SelectContent>
              </Select>
            )}
          />
        </div>
      )}

      {serverError && (
        <p role="alert" className="text-sm text-destructive">
          {serverError}
        </p>
      )}

      <SubmitButton
        isSubmitting={isSubmitting}
        submittingLabel={isEdit ? '保存中…' : '掲載中…'}
      >
        {isEdit ? '変更を保存する' : '案件を掲載する'}
      </SubmitButton>
    </form>
  )
}
