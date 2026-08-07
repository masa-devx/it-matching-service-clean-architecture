import { Hourglass } from 'lucide-react'
import type { ReviewListResult } from '@/lib/reviews'
import type { CurrentUser } from '@/lib/auth'
import { ratingLabels } from '@/lib/reviewSchema'
import { StarRating } from '@/components/StarRating'
import { ReviewForm } from '@/components/ReviewForm'

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

// レビューの表示と投稿。企業・人材で共用する。
//
// api が返す2つのフラグ（submitted / published）で3つの状態を出し分ける:
//
//   submitted=false            → 投稿フォーム
//   submitted=true,  published=false → 「相手の提出を待っています」＋自分のレビュー
//   published=true             → 双方のレビュー
//
// 配列の中身だけでは判断できない（空配列が「相手が未提出」なのか
// 「自分も未提出」なのか区別がつかない）ため、フラグで分岐する
export function ReviewSection({
  contractId,
  result,
  currentRole,
}: {
  contractId: number
  result: ReviewListResult
  currentRole: CurrentUser['role']
}) {
  // 未提出：投稿フォームを出す
  if (!result.submitted) {
    return (
      <div className="flex flex-col gap-4 rounded-lg border bg-card p-6">
        <div className="flex flex-col gap-1">
          <h3 className="font-medium">取引を評価する</h3>
          {/* 投稿前に「両者が提出するまで公開されない」ことを伝える。
              知らずに書くと「なぜすぐ表示されない？」と混乱する */}
          <p className="text-sm text-muted-foreground">
            両者が提出するまで公開されません。相手の評価を見てから書くことはできない仕組みです。
          </p>
        </div>
        <ReviewForm contractId={contractId} />
      </div>
    )
  }

  const myReview = result.reviews.find((r) => r.reviewer_role === currentRole)
  const peerReview = result.reviews.find((r) => r.reviewer_role !== currentRole)

  return (
    <div className="flex flex-col gap-4">
      {/* 提出済みだが未公開：なぜ相手のが見えないのかを説明する。
          空欄だけ見せると「バグでは？」と思われる。
          ユーザーにできることは無い（待つしかない）ので、
          「どうすればよいか」ではなく「どうなるか」の見通しを示す */}
      {!result.published && (
        <div className="flex items-start gap-3 rounded-lg border border-primary/30 bg-primary/5 p-4">
          <Hourglass
            className="mt-0.5 size-5 flex-none text-primary"
            aria-hidden="true"
          />
          <div className="flex flex-col gap-1">
            <p className="text-sm font-medium">相手の提出を待っています</p>
            <p className="text-sm text-muted-foreground">
              報復を防ぐため、両者が提出するまで互いのレビューは見えません。相手が提出すると、同時に公開されます。
            </p>
          </div>
        </div>
      )}

      {myReview && (
        <ReviewCard
          title="自分のレビュー"
          rating={myReview.rating}
          comment={myReview.comment}
          date={myReview.submitted_at}
        />
      )}

      {peerReview && (
        <ReviewCard
          title="相手からのレビュー"
          rating={peerReview.rating}
          comment={peerReview.comment}
          date={peerReview.submitted_at}
        />
      )}
    </div>
  )
}

function ReviewCard({
  title,
  rating,
  comment,
  date,
}: {
  title: string
  rating: number
  comment: string
  date: string
}) {
  return (
    <div className="flex flex-col gap-3 rounded-lg border bg-card p-6">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <h3 className="font-medium">{title}</h3>
        <p className="text-sm text-muted-foreground">
          {formatDate(date)}に提出
        </p>
      </div>
      <StarRating rating={rating} label={ratingLabels[rating]} />
      <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">
        {comment}
      </p>
    </div>
  )
}
