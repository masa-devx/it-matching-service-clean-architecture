'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { Controller, useForm } from 'react-hook-form'
import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { toast } from 'sonner'
import {
  workReportSchema,
  recentWeekStarts,
  formatWeekRange,
  type WorkReportFormValues,
  type WorkReportInputValues,
} from '@/lib/workReportSchema'
import { submitWorkReport, resubmitWorkReport } from '@/lib/contractClient'
import type { WorkReport } from '@/lib/workReports'
import { SubmitButton } from '@/components/SubmitButton'
import { FormErrorSummary } from '@/components/FormErrorSummary'
import { RequiredMark } from '@/components/RequiredMark'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const fieldLabels = {
  week_start: '対象の週',
  hours: '稼働時間',
  summary: '作業内容',
}

// report を渡すと再提出モードになる。
//
// 新規提出と再提出で違うのは「週を選べるかどうか」だけ。再提出では対象週を変えられない
// （週を変えるなら、それは別の報告になる）ため、選択欄を出さず現在の週を表示する。
// 内容の修正と再提出を1操作で行うのは api 側の設計と揃えている
// （稼働報告に下書きは無く、内容を出すこと自体が提出であるため）
export function WorkReportForm({
  contractId,
  report,
  submittedWeeks,
}: {
  contractId: number
  report?: WorkReport
  // 提出済みの週。選択肢から外して二重提出を避ける（すり抜けても API が 409）
  submittedWeeks?: string[]
}) {
  const router = useRouter()
  const [serverError, setServerError] = useState<string | null>(null)
  const [, startTransition] = useTransition()
  const isResubmit = report != null

  const {
    register,
    handleSubmit,
    control,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<WorkReportFormValues, unknown, WorkReportInputValues>({
    // 再提出でも週はフォームの値として持つ（入力欄は出さず、送信にも使わない）。
    // スキーマをモードで切り替えると useForm の型が一致しなくなるため、
    // 値を保持して検証を通す形にしている（#98 の ProjectForm の status と同じ）
    resolver: standardSchemaResolver(workReportSchema),
    defaultValues: report
      ? {
          week_start: report.week_start,
          hours: report.hours,
          summary: report.summary,
        }
      : { week_start: '', hours: 0, summary: '' },
  })

  // 未提出の週だけを選択肢にする。提出済みを残すと、選んだ瞬間に409になり
  // 「なぜ出せないのか」が分からない
  const weekOptions = recentWeekStarts().filter(
    (week) => !submittedWeeks?.includes(week),
  )

  async function onSubmit(values: WorkReportInputValues) {
    setServerError(null)

    const result = isResubmit
      ? await resubmitWorkReport(report.id, {
          hours: values.hours,
          summary: values.summary,
        })
      : await submitWorkReport(contractId, values)

    if (!result.ok) {
      setServerError(result.error)
      toast.error(result.error)
      return
    }

    toast.success(
      isResubmit ? '報告を再提出しました' : '稼働報告を提出しました',
    )
    if (!isResubmit) {
      // 続けて別の週を提出できるよう入力を空に戻す（再提出は画面が閉じるため不要）
      reset({ week_start: '', hours: 0, summary: '' })
    }
    startTransition(() => router.refresh())
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="flex flex-col gap-4"
      noValidate
    >
      <FormErrorSummary errors={errors} labels={fieldLabels} />

      {isResubmit ? (
        // 再提出では週が固定なので、入力欄ではなく表示にする
        <div className="flex flex-col gap-1">
          <span className="text-sm font-medium">対象の週</span>
          <p className="text-sm text-muted-foreground">
            {formatWeekRange(report.week_start)}
          </p>
        </div>
      ) : (
        <div className="flex flex-col gap-2">
          <Label htmlFor="week_start">
            対象の週 <RequiredMark />
          </Label>
          {/* Radix の Select は <select> ではないため register では繋がらない。
              Controller で RHF の管理下に置く */}
          <Controller
            name="week_start"
            control={control}
            render={({ field }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id="week_start" className="h-11">
                  <SelectValue placeholder="週を選択してください" />
                </SelectTrigger>
                <SelectContent>
                  {weekOptions.map((week) => (
                    <SelectItem key={week} value={week}>
                      {formatWeekRange(week)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          />
          {weekOptions.length === 0 && (
            <p className="text-xs text-muted-foreground">
              直近8週分はすべて提出済みです
            </p>
          )}
          {errors.week_start && (
            <p role="alert" className="text-sm text-destructive">
              {errors.week_start.message}
            </p>
          )}
        </div>
      )}

      <div className="flex flex-col gap-2">
        <Label htmlFor="hours">
          稼働時間（時間） <RequiredMark />
        </Label>
        <Input
          id="hours"
          type="number"
          inputMode="numeric"
          className="h-11"
          {...register('hours')}
        />
        {errors.hours && (
          <p role="alert" className="text-sm text-destructive">
            {errors.hours.message}
          </p>
        )}
      </div>

      <div className="flex flex-col gap-2">
        <Label htmlFor="summary">
          作業内容 <RequiredMark />
        </Label>
        <Textarea
          id="summary"
          rows={5}
          placeholder="実装した機能・調査した内容・打ち合わせなど"
          {...register('summary')}
        />
        <p className="text-xs text-muted-foreground">
          企業が確認する内容です。何をしたかが分かるように書いてください
        </p>
        {errors.summary && (
          <p role="alert" className="text-sm text-destructive">
            {errors.summary.message}
          </p>
        )}
      </div>

      {serverError && (
        <p role="alert" className="text-sm text-destructive">
          {serverError}
        </p>
      )}

      <SubmitButton isSubmitting={isSubmitting} submittingLabel="送信中…">
        {isResubmit ? '修正して再提出する' : '稼働報告を提出する'}
      </SubmitButton>
    </form>
  )
}
