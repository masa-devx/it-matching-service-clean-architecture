import { getProfile } from '@/lib/profile'
import { CompanyProfileForm } from '@/components/CompanyProfileForm'
import { TalentProfileForm } from '@/components/TalentProfileForm'

export const metadata = { title: 'プロフィール | Tsunagu Works' }

export default async function ProfilePage() {
  const data = await getProfile()
  if (!data) {
    return (
      <p className="text-destructive">プロフィールを取得できませんでした</p>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold">プロフィール</h1>
        <p className="text-sm text-muted-foreground">
          {data.profile
            ? '登録済みの内容を編集できます'
            : 'まだ登録されていません。入力して保存してください'}
        </p>
      </div>

      {data.role === 'company' ? (
        <CompanyProfileForm profile={data.profile} />
      ) : (
        <TalentProfileForm profile={data.profile} />
      )}
    </div>
  )
}
