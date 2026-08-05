import { getProfile } from '@/lib/profile'
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
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold">人材プロフィール</h1>
        <p className="text-sm text-muted-foreground">
          {data.profile
            ? '登録済みの内容を編集できます'
            : 'まだ登録されていません。入力して保存してください'}
        </p>
      </div>
      <TalentProfileForm profile={data.profile} />
    </div>
  )
}
