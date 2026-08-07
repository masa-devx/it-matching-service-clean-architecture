'use client'

import { useState, useTransition } from 'react'
import { useRouter } from 'next/navigation'
import { useForm, useWatch } from 'react-hook-form'
import { standardSchemaResolver } from '@hookform/resolvers/standard-schema'
import { toast } from 'sonner'
import { Star } from 'lucide-react'
import {
  reviewSchema,
  RATING_VALUES,
  ratingLabels,
  type ReviewFormValues,
  type ReviewInputValues,
} from '@/lib/reviewSchema'
import { submitReview } from '@/lib/reviewClient'
import { FormErrorSummary } from '@/components/FormErrorSummary'
import { RequiredMark } from '@/components/RequiredMark'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog'

const fieldLabels = {
  rating: '評価',
  comment: 'コメント',
}

// レビューの投稿フォーム。企業・人材で共用する。
//
// 提出は取り消せないため、送信前に確認を挟む。ただし他の確認ダイアログと違い
// 「代わりにできること」は無い（下書き保存も編集もない）ので、
// 確定することと、相手にはまだ見えないことを伝えるにとどめる
export function ReviewForm({ contractId }: { contractId: number }) {
  const router = useRouter()
  const [serverError, setServerError] = useState<string | null>(null)
  const [sending, setSending] = useState(false)
  const [, startTransition] = useTransition()

  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
  } = useForm<ReviewFormValues, unknown, ReviewInputValues>({
    resolver: standardSchemaResolver(reviewSchema),
    defaultValues: { rating: 0, comment: '' },
  })

  // 選択中の評価。星の塗り分けと、基準を示す文言の表示に使う。
  // watch ではなく useWatch を使うのは、watch がフォーム全体の変更を購読する関数を返し、
  // React Compiler がメモ化できないため（この項目だけを購読する useWatch なら問題ない）
  const rating = Number(useWatch({ control, name: 'rating' }))

  async function onSubmit(values: ReviewInputValues) {
    setServerError(null)
    setSending(true)
    const result = await submitReview(contractId, values)
    setSending(false)

    if (!result.ok) {
      setServerError(result.error)
      toast.error(result.error)
      return
    }

    toast.success('レビューを提出しました')
    startTransition(() => router.refresh())
  }

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className="flex flex-col gap-5"
      noValidate
    >
      <FormErrorSummary errors={errors} labels={fieldLabels} />

      <fieldset className="flex flex-col gap-2">
        {/* ラジオボタンのまとまりには legend で見出しを付ける
            （Label だと個々の入力を指してしまう） */}
        <legend className="text-sm font-medium">
          評価 <RequiredMark />
        </legend>

        {/* 見た目は星だが実体はラジオボタン。キーボード操作とフォーカス移動が
            標準で効き、読み上げも「ラジオボタン 5個中3個目」と伝わる */}
        <div className="flex items-center gap-1">
          {RATING_VALUES.map((value) => (
            <label
              key={value}
              className="cursor-pointer p-1 focus-within:outline focus-within:outline-2 focus-within:outline-offset-2 focus-within:outline-ring"
            >
              <input
                type="radio"
                value={value}
                className="sr-only"
                {...register('rating')}
              />
              {/* 星は装飾。意味は sr-only のテキストが担う */}
              <Star
                className={`size-7 transition-colors ${
                  value <= rating
                    ? 'fill-primary text-primary'
                    : 'text-muted-foreground/40'
                }`}
                aria-hidden="true"
              />
              <span className="sr-only">
                5段階中{value}（{ratingLabels[value]}）
              </span>
            </label>
          ))}
        </div>

        {/* 星の数だけでは基準が人によってぶれるため、選択中の意味を言葉で示す */}
        <p className="min-h-5 text-sm text-muted-foreground">
          {rating > 0 ? ratingLabels[rating] : '星を選んでください'}
        </p>

        {errors.rating && (
          <p role="alert" className="text-sm text-destructive">
            {errors.rating.message}
          </p>
        )}
      </fieldset>

      <div className="flex flex-col gap-2">
        <Label htmlFor="review-comment">
          コメント <RequiredMark />
        </Label>
        <Textarea
          id="review-comment"
          rows={5}
          placeholder="やり取りの進めやすさ、成果物の内容など"
          {...register('comment')}
        />
        <p className="text-xs text-muted-foreground">
          両者が提出すると公開され、相手にも表示されます
        </p>
        {errors.comment && (
          <p role="alert" className="text-sm text-destructive">
            {errors.comment.message}
          </p>
        )}
      </div>

      {serverError && (
        <p role="alert" className="text-sm text-destructive">
          {serverError}
        </p>
      )}

      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button type="button" className="h-11 self-start" disabled={sending}>
            {sending ? '提出中…' : 'レビューを提出する'}
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>このレビューを提出しますか？</AlertDialogTitle>
            <AlertDialogDescription>
              提出したレビューは編集も取り消しもできません。相手が提出するまでは公開されず、双方が提出した時点で同時に公開されます。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>やめる</AlertDialogCancel>
            {/* 確認を経てから実際の送信を行う。ダイアログの外のボタンは
                type="button" にしてあり、ここで初めて submit が走る */}
            <AlertDialogAction
              disabled={sending}
              onClick={handleSubmit(onSubmit)}
            >
              提出する
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </form>
  )
}
