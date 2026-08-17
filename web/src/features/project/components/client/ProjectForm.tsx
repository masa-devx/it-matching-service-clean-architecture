'use client'

import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { useState } from 'react'
import { useForm } from 'react-hook-form'

import { createProjectAction } from '../../actions/create'
import {
  projectFormSchema,
  type ProjectFormInput,
  type ProjectFormOutput,
} from '../../schemas/create'

// 数値入力の共通変換: HTML の input は常に文字列を返すため、空は undefined・それ以外は数値へ
const asNumber = {
  setValueAs: (v: unknown) => (v === '' || v == null ? undefined : Number(v)),
}

export function ProjectForm() {
  const [message, setMessage] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ProjectFormInput, unknown, ProjectFormOutput>({
    resolver: standardSchemaResolver(projectFormSchema),
    defaultValues: {
      title: '',
      description: '',
      remote_ok: false,
      required_skills: '',
    },
  })

  // handleSubmit が渡してくるのは parse 後の値（ProjectFormOutput）＝ API に送る形
  const onSubmit = async (data: ProjectFormOutput) => {
    setMessage(null)
    const result = await createProjectAction(data)
    setMessage(
      result.ok
        ? `案件を作成しました（id: ${result.project.id}）`
        : result.error,
    )
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="max-w-xl space-y-4">
      <div>
        <label htmlFor="title" className="block font-medium">
          タイトル
        </label>
        <input
          id="title"
          {...register('title')}
          className="w-full rounded border p-2"
        />
        {errors.title && (
          <p role="alert" className="text-sm text-red-600">
            {errors.title.message}
          </p>
        )}
      </div>

      <div>
        <label htmlFor="description" className="block font-medium">
          詳細
        </label>
        <textarea
          id="description"
          {...register('description')}
          className="w-full rounded border p-2"
        />
        {errors.description && (
          <p role="alert" className="text-sm text-red-600">
            {errors.description.message}
          </p>
        )}
      </div>

      <div className="flex gap-4">
        <div>
          <label htmlFor="hourly_rate_min" className="block font-medium">
            時給下限（円・任意）
          </label>
          <input
            id="hourly_rate_min"
            type="number"
            {...register('hourly_rate_min', asNumber)}
            className="w-full rounded border p-2"
          />
          {errors.hourly_rate_min && (
            <p role="alert" className="text-sm text-red-600">
              {errors.hourly_rate_min.message}
            </p>
          )}
        </div>
        <div>
          <label htmlFor="hourly_rate_max" className="block font-medium">
            時給上限（円・任意）
          </label>
          <input
            id="hourly_rate_max"
            type="number"
            {...register('hourly_rate_max', asNumber)}
            className="w-full rounded border p-2"
          />
        </div>
      </div>

      <div>
        <label htmlFor="hours_per_week" className="block font-medium">
          週の稼働時間
        </label>
        <input
          id="hours_per_week"
          type="number"
          {...register('hours_per_week', asNumber)}
          className="w-full rounded border p-2"
        />
        {errors.hours_per_week && (
          <p role="alert" className="text-sm text-red-600">
            {errors.hours_per_week.message}
          </p>
        )}
      </div>

      <div>
        <label className="flex items-center gap-2">
          <input type="checkbox" {...register('remote_ok')} />
          リモート可
        </label>
      </div>

      <div>
        <label htmlFor="required_skills" className="block font-medium">
          必須スキル（カンマ区切り）
        </label>
        <input
          id="required_skills"
          placeholder="Go, PostgreSQL"
          {...register('required_skills')}
          className="w-full rounded border p-2"
        />
      </div>

      <button
        type="submit"
        disabled={isSubmitting}
        className="rounded bg-blue-600 px-4 py-2 text-white disabled:opacity-50"
      >
        {isSubmitting ? '作成中…' : '案件を作成'}
      </button>

      {message && <p className="font-medium">{message}</p>}
    </form>
  )
}
