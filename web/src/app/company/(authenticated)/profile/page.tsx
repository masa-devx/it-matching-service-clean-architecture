import { getProfile } from '@/lib/profile'
import { PageHeader } from '@/components/PageHeader'
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
      <PageHeader
        title="企業プロフィール"
        description={
          data.profile
            ? '登録済みの内容を編集できます'
            : 'まだ登録されていません。入力して保存してください'
        }
      />
      <CompanyProfileForm profile={data.profile} />
    </div>
  )
}
