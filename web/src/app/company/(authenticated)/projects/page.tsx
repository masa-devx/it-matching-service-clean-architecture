export const metadata = { title: '案件管理 | Tsunagu Works' }

// 自社が掲載した案件の一覧・管理。中身は #10 で実装する
export default function CompanyProjectsPage() {
  return (
    <div className="flex flex-col gap-2">
      <h1 className="text-2xl font-bold">案件管理</h1>
      <p className="text-muted-foreground">
        自社が掲載した案件の一覧です（準備中・#10 で実装）
      </p>
    </div>
  )
}
