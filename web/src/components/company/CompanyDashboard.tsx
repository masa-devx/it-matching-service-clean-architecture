import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { StatCard } from '@/components/StatCard'
import { PageHeader } from '@/components/PageHeader'
import { ProfileIncompleteNotice } from '@/components/ProfileIncompleteNotice'

export function CompanyDashboard({
  hasProfile,
  publishedCount,
  activeContractCount,
  pendingReportCount,
}: {
  hasProfile: boolean
  publishedCount: number
  activeContractCount: number
  pendingReportCount: number
}) {
  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="案件を掲載して人材を探す"
        description="掲載した案件は人材ユーザーの一覧に表示されます"
      />

      {/* 企業プロフィールは案件掲載の前提条件（API側で400になる）ため強調して案内する */}
      {!hasProfile && (
        <ProfileIncompleteNotice
          required
          href="/company/profile"
          message="案件を掲載するには、先に企業プロフィールの登録が必要です"
        />
      )}

      <div className="flex flex-wrap gap-3">
        {/* asChild は中身を <a> に差し替えるため disabled が効かない。
            リンクにせず無効なボタンとして描画し、押せないことを明示する */}
        {hasProfile ? (
          <Button asChild className="h-11">
            <Link href="/company/projects">案件を掲載する</Link>
          </Button>
        ) : (
          <Button className="h-11" disabled>
            案件を掲載する
          </Button>
        )}
        <Button asChild variant="outline" className="h-11">
          <Link href="/company/projects">掲載中の案件を管理</Link>
        </Button>
      </div>

      {/* 未確認の稼働報告は「相手を待たせている」状態なので、件数だけでなく導線も出す。
          0件のときは何も出さない（対応が不要なのに通知があると、次から見なくなる） */}
      {pendingReportCount > 0 && (
        <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-primary/30 bg-primary/5 p-4">
          <p className="text-sm">
            <span className="font-bold tabular-nums text-primary">
              {pendingReportCount}件
            </span>
            の稼働報告が未確認です
          </p>
          <Button asChild className="h-11">
            <Link href="/company/contracts">契約を確認する</Link>
          </Button>
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard label="稼働中の契約" value={activeContractCount} />
        <StatCard label="公開中の案件（全体）" value={publishedCount} />
      </div>
    </div>
  )
}
