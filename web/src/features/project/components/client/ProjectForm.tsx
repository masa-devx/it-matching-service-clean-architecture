'use client'

import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import type { TsunaguWorksProject } from '@repo/api-client/company/generated/models'
import { useState } from 'react'
import { Controller, useForm } from 'react-hook-form'

import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

import { createProjectAction } from '../../actions/create'
import { updateProjectAction } from '../../actions/update'
import {
  projectFormSchema,
  type ProjectFormInput,
  type ProjectFormOutput,
} from '../../schemas/create'

// 数値入力の共通変換: HTML の input は常に文字列を返すため、空は undefined・それ以外は数値へ
const asNumber = {
  setValueAs: (v: unknown) => (v === '' || v == null ? undefined : Number(v)),
}

// 作成と編集で同じフォームを使う。project の有無がモード:
// 無し=作成（空の初期値・成功時はメッセージ表示）／有り=編集（値を復元・成功時は一覧へ redirect）。
// API 側も作成と編集で同じ入力形（ProjectUpdateInput は CreateInput の spread）なので、スキーマも1本で足りる
export function ProjectForm({ project }: { project?: TsunaguWorksProject }) {
  const [message, setMessage] = useState<string | null>(null)

  // 作成モードは数値項目が未入力（undefined）で始まる部分値のため、
  // ProjectFormInput 全体では注釈しない（useForm の DefaultValues が部分値を許容する）
  const defaultValues = project
    ? {
        title: project.title,
        description: project.description,
        hourly_rate_min: project.hourly_rate_min ?? undefined,
        hourly_rate_max: project.hourly_rate_max ?? undefined,
        hours_per_week: project.hours_per_week,
        remote_ok: project.remote_ok,
        // フォーム上はカンマ区切りの1テキスト（schemas/create.ts の transform と対）
        required_skills: project.required_skills.join(', '),
      }
    : {
        title: '',
        description: '',
        remote_ok: false,
        required_skills: '',
      }

  const {
    register,
    control,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ProjectFormInput, unknown, ProjectFormOutput>({
    resolver: standardSchemaResolver(projectFormSchema),
    defaultValues,
  })

  // handleSubmit が渡してくるのは parse 後の値（ProjectFormOutput）＝ API に送る形
  const onSubmit = async (data: ProjectFormOutput) => {
    setMessage(null)
    if (project) {
      // 成功時は action 内の redirect で一覧へ遷移するため、戻り値があるのは失敗時だけ
      const result = await updateProjectAction(project.id, data)
      if (result?.error) {
        setMessage(result.error)
      }
      return
    }
    const result = await createProjectAction(data)
    setMessage(
      result.ok
        ? `案件を作成しました（id: ${result.project.id}）`
        : result.error,
    )
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="max-w-xl space-y-4">
      <div className="space-y-1.5">
        <Label htmlFor="title">タイトル</Label>
        <Input
          id="title"
          aria-invalid={errors.title ? true : undefined}
          {...register('title')}
        />
        {errors.title && (
          <p role="alert" className="text-sm text-destructive">
            {errors.title.message}
          </p>
        )}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="description">詳細</Label>
        <Textarea
          id="description"
          aria-invalid={errors.description ? true : undefined}
          {...register('description')}
        />
        {errors.description && (
          <p role="alert" className="text-sm text-destructive">
            {errors.description.message}
          </p>
        )}
      </div>

      <div className="flex gap-4">
        <div className="space-y-1.5">
          <Label htmlFor="hourly_rate_min">時給下限（円・任意）</Label>
          <Input
            id="hourly_rate_min"
            type="number"
            aria-invalid={errors.hourly_rate_min ? true : undefined}
            {...register('hourly_rate_min', asNumber)}
          />
          {errors.hourly_rate_min && (
            <p role="alert" className="text-sm text-destructive">
              {errors.hourly_rate_min.message}
            </p>
          )}
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="hourly_rate_max">時給上限（円・任意）</Label>
          <Input
            id="hourly_rate_max"
            type="number"
            {...register('hourly_rate_max', asNumber)}
          />
        </div>
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="hours_per_week">週の稼働時間</Label>
        <Input
          id="hours_per_week"
          type="number"
          aria-invalid={errors.hours_per_week ? true : undefined}
          {...register('hours_per_week', asNumber)}
        />
        {errors.hours_per_week && (
          <p role="alert" className="text-sm text-destructive">
            {errors.hours_per_week.message}
          </p>
        )}
      </div>

      {/* Radix の Checkbox は native input ではないため register が使えず、Controller で値を接続する */}
      <Controller
        control={control}
        name="remote_ok"
        render={({ field }) => (
          <div className="flex items-center gap-2">
            <Checkbox
              id="remote_ok"
              checked={field.value}
              onCheckedChange={(checked) => field.onChange(checked === true)}
            />
            <Label htmlFor="remote_ok">リモート可</Label>
          </div>
        )}
      />

      <div className="space-y-1.5">
        <Label htmlFor="required_skills">必須スキル（カンマ区切り）</Label>
        <Input
          id="required_skills"
          placeholder="Go, PostgreSQL"
          {...register('required_skills')}
        />
      </div>

      <Button type="submit" disabled={isSubmitting}>
        {project
          ? isSubmitting
            ? '保存中…'
            : '保存する'
          : isSubmitting
            ? '作成中…'
            : '案件を作成'}
      </Button>

      {message && (
        <p role="alert" className="font-medium">
          {message}
        </p>
      )}
    </form>
  )
}
