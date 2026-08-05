export const metadata = { title: '案件を探す | Tsunagu Works' }

// 公開中の案件一覧。中身は #10 で実装する
export default function TalentProjectsPage() {
  return (
    <div className="flex flex-col gap-2">
      <h1 className="text-2xl font-bold">案件を探す</h1>
      <p className="text-muted-foreground">
        公開中の案件一覧です（準備中・#10 で実装）
      </p>
    </div>
  )
}
