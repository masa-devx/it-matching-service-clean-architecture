import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { StatCard } from '@/components/StatCard'
import { ProfileIncompleteNotice } from '@/components/ProfileIncompleteNotice'

export function CompanyDashboard({
  hasProfile,
  publishedCount,
}: {
  hasProfile: boolean
  publishedCount: number
}) {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold">案件を掲載して人材を探す</h1>
        <p className="text-sm text-muted-foreground">
          掲載した案件は人材ユーザーの一覧に表示されます
        </p>
      </div>

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

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard label="公開中の案件（全体）" value={publishedCount} />
      </div>
    </div>
  )
}
