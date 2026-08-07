import { getProfile } from '@/lib/profile'
import { countPublishedProjects } from '@/lib/projects'
import { countActiveContracts } from '@/lib/contracts'
import { TalentDashboard } from '@/components/talent/TalentDashboard'

export const metadata = { title: 'ダッシュボード | Tsunagu Works' }

export default async function TalentDashboardPage() {
  const [profile, publishedCount, activeContractCount] = await Promise.all([
    getProfile(),
    countPublishedProjects(),
    countActiveContracts(),
  ])

  return (
    <TalentDashboard
      hasProfile={profile?.profile != null}
      publishedCount={publishedCount}
      activeContractCount={activeContractCount}
    />
  )
}
