import { ProjectForm } from './ProjectForm'

// Phase 0 の検証ページ（生成チェーンの疎通確認用・認証ガード外）。
// features / external 構成が入る Phase 1 以降で正式な配置へ移設する
export default function Page() {
  return (
    <main className="mx-auto max-w-2xl p-8">
      <h1 className="mb-6 text-2xl font-bold">案件を作成（Phase 0 検証）</h1>
      <ProjectForm />
    </main>
  )
}
