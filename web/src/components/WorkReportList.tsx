import { CalendarRange } from 'lucide-react'
import type { WorkReport } from '@/lib/workReports'
import { formatWeekRange } from '@/lib/workReportSchema'
import {
  WorkReportStatusBadge,
  workReportStatusDescription,
} from '@/components/WorkReportStatusBadge'
import { EmptyState } from '@/components/EmptyState'

// 稼働報告の一覧。企業・人材のどちらの画面でも同じものを表示する
// （共有された記録なので、視点で見え方を変えない）。
// 操作ボタンだけがロールで違うため、actions を差し込めるようにしている
export function WorkReportList({
  reports,
  emptyMessage,
  renderActions,
}: {
  reports: WorkReport[]
  emptyMessage: string
  renderActions?: (report: WorkReport) => React.ReactNode
}) {
  if (reports.length === 0) {
    return (
      <EmptyState
        icon={CalendarRange}
        title="稼働報告がありません"
        description={emptyMessage}
      />
    )
  }

  return (
    <ul className="flex flex-col gap-3">
      {reports.map((report) => (
        <li
          key={report.id}
          className="flex flex-col gap-3 rounded-lg border bg-card p-6"
        >
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div className="flex flex-col gap-1">
              {/* 開始日だけだと「その週」だと分かりにくいので範囲で見せる */}
              <p className="font-bold">{formatWeekRange(report.week_start)}</p>
              <p className="text-sm text-muted-foreground tabular-nums">
                稼働 {report.hours}時間
              </p>
            </div>
            <WorkReportStatusBadge status={report.status} />
          </div>

          <p className="text-sm text-muted-foreground">
            {workReportStatusDescription(report.status)}
          </p>

          <div className="flex flex-col gap-1">
            <h3 className="text-sm font-medium">作業内容</h3>
            <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">
              {report.summary}
            </p>
          </div>

          {/* 差し戻し理由は、承認済み・確認待ちでは空になる（api 側で消している） */}
          {report.review_note && (
            <div className="flex flex-col gap-1 rounded-md border border-destructive/30 bg-destructive/5 p-4">
              <h3 className="text-sm font-medium text-destructive">
                差し戻しの理由
              </h3>
              <p className="whitespace-pre-wrap break-words text-sm leading-relaxed">
                {report.review_note}
              </p>
            </div>
          )}

          {renderActions?.(report)}
        </li>
      ))}
    </ul>
  )
}
