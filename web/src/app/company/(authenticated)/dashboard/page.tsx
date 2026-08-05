import { getProfile } from '@/lib/profile'
import { getPublishedProjects } from '@/lib/projects'
import { CompanyDashboard } from '@/components/company/CompanyDashboard'

export const metadata = { title: 'ダッシュボード | Tsunagu Works' }

export default async function CompanyDashboardPage() {
  // 独立した取得なので直列に待たず並行実行する
  const [profile, projects] = await Promise.all([
    getProfile(),
    getPublishedProjects(100),
  ])

  return (
    <CompanyDashboard
      hasProfile={profile?.profile != null}
      publishedCount={projects?.length ?? 0}
    />
  )
}
