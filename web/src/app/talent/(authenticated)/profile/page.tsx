import { getProfile } from '@/lib/profile'
import { PageHeader } from '@/components/PageHeader'
import { TalentProfileForm } from '@/components/talent/TalentProfileForm'

export const metadata = { title: 'プロフィール | Tsunagu Works' }

// ロールは layout のガードで talent に確定しているため、ここでの分岐は不要
export default async function TalentProfilePage() {
  const data = await getProfile()
  if (!data || data.role !== 'talent') {
    return (
      <p className="text-destructive">プロフィールを取得できませんでした</p>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="人材プロフィール"
        description={
          data.profile
            ? '登録済みの内容を編集できます'
            : 'まだ登録されていません。入力して保存してください'
        }
      />
      <TalentProfileForm profile={data.profile} />
    </div>
  )
}
