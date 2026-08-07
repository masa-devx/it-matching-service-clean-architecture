import Link from 'next/link'
import { notFound } from 'next/navigation'
import { ArrowLeft, Banknote, Clock, MapPin } from 'lucide-react'
import { getMyContract } from '@/lib/contracts'
import { getWorkReports } from '@/lib/workReports'
import {
  ContractStatusBadge,
  contractStatusDescription,
} from '@/components/ContractStatusBadge'
import { WorkReportList } from '@/components/WorkReportList'
import { ContractActions } from '@/components/talent/ContractActions'
import { WorkReportForm } from '@/components/talent/WorkReportForm'
import { WorkReportEditor } from '@/components/talent/WorkReportEditor'
import { Separator } from '@/components/ui/separator'

export const metadata = { title: '契約の詳細 | Tsunagu Works' }

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  })
}

export default async function TalentContractDetailPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = await params
  const contractId = Number(id)

  // 契約と稼働報告は別のエンドポイント。独立した取得なので並行実行する
  const [contract, reportResult] = await Promise.all([
    getMyContract(contractId),
    getWorkReports(contractId),
  ])
  // 当事者でない契約・存在しない契約は API が404を返す
  if (!contract) {
    notFound()
  }

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
        href="/talent/contracts"
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
        <p className="text-muted-foreground">{contract.company_name}</p>
      </header>

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

      {/* 現在の状態で実行できる操作だけが並ぶ（遷移表の写し） */}
      <ContractActions contractId={contract.id} status={contract.status} />

      {/* 条件が案件と違って見えることがあるため、固定されている事実を明示する */}
      <p className="text-sm text-muted-foreground">
        この契約の条件は、承諾した時点の内容で固定されています（案件の掲載内容が変更されても影響しません）。
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
          <p className="text-sm tabular-nums text-muted-foreground">
            {contract.work_report_count}件
          </p>
        </div>

        {reportResult === null ? (
          <p className="text-destructive">稼働報告を取得できませんでした</p>
        ) : (
          <>
            {/* 提出できるのは稼働中の契約だけ（api 側も409で拒否する）。
                検収待ち・完了後にフォームを出すと、押せるのに失敗するボタンになる */}
            {contract.status === 'working' && (
              <div className="flex flex-col gap-4 rounded-lg border bg-card p-6">
                <h3 className="font-medium">今週の稼働を報告する</h3>
                <WorkReportForm
                  contractId={contract.id}
                  submittedWeeks={reportResult.work_reports.map(
                    (r) => r.week_start,
                  )}
                />
              </div>
            )}

            <WorkReportList
              reports={reportResult.work_reports}
              emptyMessage="稼働中になると、週ごとに報告を提出できます"
              // 差し戻された報告だけ修正できる。承認済み・確認待ちは触れない
              renderActions={(report) =>
                report.status === 'rejected' ? (
                  <WorkReportEditor report={report} />
                ) : null
              }
            />
          </>
        )}
      </section>
    </article>
  )
}
