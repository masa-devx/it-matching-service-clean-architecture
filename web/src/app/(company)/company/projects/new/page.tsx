import { ProjectForm } from '@/features/project/components/client/ProjectForm'

export default function Page() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">案件を作成</h1>
      <ProjectForm />
    </div>
  )
}
