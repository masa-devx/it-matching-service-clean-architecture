import { getProfile } from '@/lib/profile'
import { countPublishedProjects } from '@/lib/projects'
import { TalentDashboard } from '@/components/talent/TalentDashboard'

export const metadata = { title: 'ダッシュボード | Tsunagu Works' }

export default async function TalentDashboardPage() {
  const [profile, publishedCount] = await Promise.all([
    getProfile(),
    countPublishedProjects(),
  ])

  return (
    <TalentDashboard
      hasProfile={profile?.profile != null}
      publishedCount={publishedCount}
    />
  )
}
