import Link from 'next/link'
import { notFound } from 'next/navigation'
import { ArrowLeft, Banknote, Clock, MapPin } from 'lucide-react'
import { getMyContract } from '@/lib/contracts'
import { getWorkReports } from '@/lib/workReports'
import { getMessages } from '@/lib/messages'
import { getReviews } from '@/lib/reviews'
import {
  ContractStatusBadge,
  contractStatusDescription,
} from '@/components/ContractStatusBadge'
import { WorkReportList } from '@/components/WorkReportList'
import { CompanyContractActions } from '@/components/company/CompanyContractActions'
import { WorkReportReviewActions } from '@/components/company/WorkReportReviewActions'
import { ReviewSection } from '@/components/ReviewSection'
import { MessageThread } from '@/components/MessageThread'
import { MessageForm } from '@/components/MessageForm'
import { Separator } from '@/components/ui/separator'

export const metadata = { title: '契約の詳細 | Tsunagu Works' }

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

export default async function CompanyContractDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const contractId = Number(id)

  // 契約と稼働報告は別のエンドポイント。独立した取得なので並行実行する
  const [contract, reportResult, messageResult, reviewResult] =
    await Promise.all([
      getMyContract(contractId),
      getWorkReports(contractId),
      getMessages(contractId),
      getReviews(contractId),
    ])
  // 当事者でない契約・存在しない契約は API が404を返す
  if (!contract) {
    notFound()
  }

  // 進行中の契約だけメッセージを送れる（api/messages.go の messageSendableStatuses と対応）
  const canSendMessage = ['active', 'working', 'reviewing'].includes(
    contract.status,
  )

  const conditions = [
    {
      icon: Banknote,
      label: '時給',
      value: `${contract.hourly_rate.toLocaleString()}円`,
    },
    {
      icon: Clock,
      label: '週の稼働時間',
      value: `${contract.hours_per_week}時間`,
    },
    {
      icon: MapPin,
      label: '勤務形態',
      value: contract.remote_ok ? 'フルリモート可' : '出社あり',
    },
  ]

  return (
    <article className="flex flex-col gap-6">
      <Link
        href="/company/contracts"
        className="flex w-fit items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="size-4" aria-hidden="true" />
        契約一覧に戻る
      </Link>

      <header className="flex flex-col gap-3">
        <div className="flex flex-wrap items-center gap-2">
          <ContractStatusBadge status={contract.status} />
          <span className="text-sm text-muted-foreground">
            {contractStatusDescription(contract.status)}
          </span>
        </div>
        <h1 className="text-2xl font-bold leading-snug tracking-tight">
          {contract.title}
        </h1>
        <p className="text-muted-foreground">{contract.talent_name}</p>
      </header>

      {/* 現在の状態で実行できる操作だけが並ぶ（遷移表の写し） */}
      <CompanyContractActions
        contractId={contract.id}
        status={contract.status}
      />

      <dl className="grid gap-4 rounded-lg border bg-card p-6 sm:grid-cols-3">
        {conditions.map(({ icon: Icon, label, value }) => (
          <div key={label} className="flex items-start gap-3">
            <Icon className="mt-0.5 size-5 text-primary" aria-hidden="true" />
            <div className="flex flex-col gap-0.5">
              <dt className="text-sm text-muted-foreground">{label}</dt>
              <dd className="font-bold tabular-nums">{value}</dd>
            </div>
          </div>
        ))}
      </dl>

      {/* 案件を編集しても契約の条件は変わらない。企業側にも明示しておく */}
      <p className="text-sm text-muted-foreground">
        この契約の条件は、人材が承諾した時点の内容で固定されています（案件の掲載内容を変更しても影響しません）。
      </p>

      {(contract.started_at || contract.completed_at) && (
        <dl className="flex flex-wrap gap-x-8 gap-y-2 text-sm">
          {contract.started_at && (
            <div className="flex gap-2">
              <dt className="text-muted-foreground">稼働開始</dt>
              <dd>{formatDate(contract.started_at)}</dd>
            </div>
          )}
          {contract.completed_at && (
            <div className="flex gap-2">
              <dt className="text-muted-foreground">完了</dt>
              <dd>{formatDate(contract.completed_at)}</dd>
            </div>
          )}
        </dl>
      )}

      <Separator />

      <section className="flex flex-col gap-4">
        <div className="flex flex-wrap items-baseline justify-between gap-2">
          <h2 className="font-bold">稼働報告</h2>
          <p className="text-sm tabular-nums">
            <span className="text-muted-foreground">
              {contract.work_report_count}件
            </span>
            {contract.pending_report_count > 0 && (
              <span className="ml-1.5 font-medium text-primary">
                （{contract.pending_report_count}件が未確認）
              </span>
            )}
          </p>
        </div>

        {reportResult === null ? (
          <p className="text-destructive">稼働報告を取得できませんでした</p>
        ) : (
          <WorkReportList
            reports={reportResult.work_reports}
            emptyMessage="人材が稼働を開始すると、週ごとの報告が届きます"
            // 確認待ちの報告だけ操作できる（部品側で判定している）
            renderActions={(report) => (
              <WorkReportReviewActions report={report} />
            )}
          />
        )}
      </section>

      {/* レビューは完了した契約にだけ表示する。
          進行中に評価を書けると「悪い評価をつけられたくないから検収を通す」という
          圧力が生まれ、検収の中立性が失われる（api も409で拒否する） */}
      {contract.status === 'completed' && reviewResult && (
        <>
          <Separator />
          <section className="flex flex-col gap-4">
            <h2 className="font-bold">レビュー</h2>
            <ReviewSection
              contractId={contract.id}
              result={reviewResult}
              currentRole="company"
            />
          </section>
        </>
      )}

      <Separator />

      {/* メッセージは契約の文脈で交わされるため、別ページにせず詳細に置く
          （条件や稼働報告を見ながら書けるようにする） */}
      <section className="flex flex-col gap-4">
        <h2 className="font-bold">メッセージ</h2>

        {messageResult === null ? (
          <p className="text-destructive">メッセージを取得できませんでした</p>
        ) : (
          <MessageThread
            messages={messageResult.messages}
            currentRole="company"
          />
        )}

        {/* 終了した契約には送信できない（api も409で拒否する）。
            押せるのに必ず失敗するフォームは出さない */}
        {canSendMessage && <MessageForm contractId={contract.id} />}
      </section>
    </article>
  )
}
