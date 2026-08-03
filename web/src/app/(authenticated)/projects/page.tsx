export const metadata = { title: '案件一覧 | Tsunagu Works' }

// 認証ガード(#8)の確認用プレースホルダ。一覧の実装は #10 で行う
export default function ProjectsPage() {
  return (
    <div className="flex flex-col gap-2">
      <h1 className="text-2xl font-bold">案件一覧</h1>
      <p className="text-muted-foreground">準備中です（#10 で実装予定）</p>
    </div>
  )
}
