import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { StatCard } from '@/components/StatCard'
import { PageHeader } from '@/components/PageHeader'
import { ProfileIncompleteNotice } from '@/components/ProfileIncompleteNotice'

export function TalentDashboard({
  hasProfile,
  publishedCount,
}: {
  hasProfile: boolean
  publishedCount: number
}) {
  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="案件を探して応募する"
        description="スキルや稼働条件に合う案件を探せます"
      />

      {/* 人材プロフィールは応募の必須条件ではないが、登録した方が有利という案内に留める */}
      {!hasProfile && (
        <ProfileIncompleteNotice
          href="/talent/profile"
          message="プロフィールを登録すると、企業に見つけてもらいやすくなります"
        />
      )}

      <div className="flex flex-wrap gap-3">
        <Button asChild className="h-11">
          <Link href="/talent/projects">案件を探す</Link>
        </Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard label="公開中の案件" value={publishedCount} />
      </div>
    </div>
  )
}
