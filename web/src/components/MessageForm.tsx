'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { toast } from 'sonner'
import { messageSchema, type MessageFormValues } from '@/lib/messageSchema'
import { sendMessage } from '@/lib/messageClient'
import { SubmitButton } from '@/components/SubmitButton'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

// メッセージの送信フォーム。企業・人材で共用する。
//
// 送信前に「連絡先は伏せられる」ことを伝えておく。送ってから伏せられるより、
// 書く前に分かっているほうがユーザーの体験としてよい
// （書き直しの手間が減り、「勝手に消された」という印象にもならない）
export function MessageForm({ contractId }: { contractId: number }) {
  const router = useRouter()
  const [serverError, setServerError] = useState<string | null>(null)
  const [, startTransition] = useTransition()

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<MessageFormValues>({
    resolver: standardSchemaResolver(messageSchema),
    defaultValues: { body: '' },
  })

  async function onSubmit(values: MessageFormValues) {
    setServerError(null)

    const result = await sendMessage(contractId, values.body)
    if (!result.ok) {
      setServerError(result.error)
      toast.error(result.error)
      return
    }

    // 続けて書けるよう入力を空に戻す。トーストは出さない——
    // 送信結果は会話に自分の発言が増えることで分かるため、通知は過剰になる
    reset({ body: '' })
    startTransition(() => router.refresh())
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="flex flex-col gap-3"
      noValidate
    >
      <div className="flex flex-col gap-2">
        <Label htmlFor="message-body">メッセージ</Label>
        <Textarea
          id="message-body"
          rows={3}
          placeholder="作業の相談や進捗の共有に使えます"
          {...register('body')}
        />
        <p className="text-xs text-muted-foreground">
          メールアドレス・電話番号・URLは自動的に伏せられます（安全な取引のため）
        </p>
        {errors.body && (
          <p role="alert" className="text-sm text-destructive">
            {errors.body.message}
          </p>
        )}
      </div>

      {serverError && (
        <p role="alert" className="text-sm text-destructive">
          {serverError}
        </p>
      )}

      <SubmitButton isSubmitting={isSubmitting} submittingLabel="送信中…">
        送信する
      </SubmitButton>
    </form>
  )
}
