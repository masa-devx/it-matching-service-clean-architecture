'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { toast } from 'sonner'
import {
  applicationSchema,
  type ApplicationInput,
} from '@/lib/applicationSchema'
import { applyToProject } from '@/lib/applicationClient'
import { SubmitButton } from '@/components/SubmitButton'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { RequiredMark } from '@/components/RequiredMark'

export function ApplyForm({ projectId }: { projectId: number }) {
  const router = useRouter()
  const [serverError, setServerError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ApplicationInput>({
    resolver: standardSchemaResolver(applicationSchema),
    defaultValues: { message: '' },
  })

  async function onSubmit(values: ApplicationInput) {
    setServerError(null)

    const result = await applyToProject(projectId, values.message)
    if (!result.ok) {
      setServerError(result.error)
      toast.error(result.error)
      return
    }

    toast.success('応募しました')
    // 応募後は「応募済み」表示に切り替わる。サーバー側の状態を取り直す
    router.refresh()
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="flex flex-col gap-4"
      noValidate
    >
      <div className="flex flex-col gap-2">
        <Label htmlFor="message">
          志望動機 <RequiredMark />
        </Label>
        <Textarea
          id="message"
          rows={6}
          placeholder="この案件で活かせる経験や、興味を持った理由を書いてください"
          {...register('message')}
        />
        <p className="text-xs text-muted-foreground">
          企業の応募者一覧に表示されます（2000文字以内）
        </p>
        {errors.message && (
          <p role="alert" className="text-sm text-destructive">
            {errors.message.message}
          </p>
        )}
      </div>

      {serverError && (
        <p role="alert" className="text-sm text-destructive">
          {serverError}
        </p>
      )}

      <SubmitButton isSubmitting={isSubmitting} submittingLabel="送信中…">
        応募する
      </SubmitButton>
    </form>
  )
}
