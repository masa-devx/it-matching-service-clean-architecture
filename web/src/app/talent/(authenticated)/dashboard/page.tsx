import { getProfile } from '@/lib/profile'
import { getPublishedProjects } from '@/lib/projects'
import { TalentDashboard } from '@/components/talent/TalentDashboard'

export const metadata = { title: 'ダッシュボード | Tsunagu Works' }

export default async function TalentDashboardPage() {
  const [profile, projects] = await Promise.all([
    getProfile(),
    getPublishedProjects(100),
  ])

  return (
    <TalentDashboard
      hasProfile={profile?.profile != null}
      publishedCount={projects?.length ?? 0}
    />
  )
}
