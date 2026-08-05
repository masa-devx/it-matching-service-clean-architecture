import Link from 'next/link'
import { redirect } from 'next/navigation'
import { getProfile } from '@/lib/profile'
import { ProjectForm } from '@/components/company/ProjectForm'

export const metadata = { title: '新規掲載 | Tsunagu Works' }

export default async function NewProjectPage() {
  const data = await getProfile()
  // 掲載には企業プロフィールが必要（API側も400で弾く）。先に登録画面へ誘導する
  if (data?.role === 'company' && !data.profile) {
    redirect('/company/profile')
  }

  return (
    <div className="flex flex-col gap-6">
      <Link
        href="/company/projects"
        className="text-sm text-muted-foreground hover:text-foreground"
      >
        ← 案件管理に戻る
      </Link>

      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-bold">案件を掲載する</h1>
        <p className="text-sm text-muted-foreground">
          公開すると人材ユーザーの一覧に表示されます
        </p>
      </div>

      <ProjectForm />
    </div>
  )
}
