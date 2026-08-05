import { getProfile } from '@/lib/profile'
import { CompanyProfileForm } from '@/components/company/CompanyProfileForm'

export const metadata = { title: 'プロフィール | Tsunagu Works' }

// ロールは layout のガードで company に確定しているため、ここでの分岐は不要
export default async function CompanyProfilePage() {
  const data = await getProfile()
  if (!data || data.role !== 'company') {
    return (
      <p className="text-destructive">プロフィールを取得できませんでした</p>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold">企業プロフィール</h1>
        <p className="text-sm text-muted-foreground">
          {data.profile
            ? '登録済みの内容を編集できます'
            : 'まだ登録されていません。入力して保存してください'}
        </p>
      </div>
      <CompanyProfileForm profile={data.profile} />
    </div>
  )
}
