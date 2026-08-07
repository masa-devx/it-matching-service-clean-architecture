import { getProfile } from '@/lib/profile'
import { countPublishedProjects } from '@/lib/projects'
import { countActiveContracts, countPendingWorkReports } from '@/lib/contracts'
import { CompanyDashboard } from '@/components/company/CompanyDashboard'

export const metadata = { title: 'ダッシュボード | Tsunagu Works' }

export default async function CompanyDashboardPage() {
  // 独立した取得なので直列に待たず並行実行する
  const [profile, publishedCount, activeContractCount, pendingReportCount] =
    await Promise.all([
      getProfile(),
      countPublishedProjects(),
      countActiveContracts(),
      countPendingWorkReports(),
    ])

  return (
    <CompanyDashboard
      hasProfile={profile?.profile != null}
      publishedCount={publishedCount}
      activeContractCount={activeContractCount}
      pendingReportCount={pendingReportCount}
    />
  )
}
