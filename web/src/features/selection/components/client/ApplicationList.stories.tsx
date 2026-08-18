import type { Meta, StoryObj } from '@storybook/nextjs-vite'
import type { TsunaguWorksApplicationForCompany } from '@repo/api-client/company/generated/models'

import { fetchApplications } from '../../actions/read.mock'
import { ApplicationList } from './ApplicationList'

// ページ級 story: /company/projects/{id}/applications の実体（選考画面）。
// talent 側（Pages/応募一覧）と同じ応募が、視点で違うラベル・操作を持つ対比を確認できる
const meta = {
  title: 'Pages/選考（company）',
  component: ApplicationList,
  args: { projectId: 1 },
} satisfies Meta<typeof ApplicationList>

export default meta
type Story = StoryObj<typeof meta>

const applications: TsunaguWorksApplicationForCompany[] = [
  {
    id: 3,
    project_id: 1,
    status: 'applied',
    message:
      'バックエンドの経験を活かせると思い応募しました。\nよろしくお願いします。',
    talent_display_name: '山田太郎',
    talent_skills: ['Go', 'PostgreSQL', 'TypeScript'],
    created_at: '2026-08-16T09:00:00Z',
  },
  {
    id: 2,
    project_id: 1,
    status: 'offered',
    message: 'フロントエンドが得意です。',
    talent_display_name: '鈴木花子',
    talent_skills: ['React', 'Next.js'],
    created_at: '2026-08-14T09:00:00Z',
  },
  {
    id: 1,
    project_id: 1,
    status: 'withdrawn',
    message: '',
    talent_display_name: '佐藤次郎',
    talent_skills: ['AWS'],
    created_at: '2026-08-10T09:00:00Z',
  },
]

export const WithData: Story = {
  name: 'データあり',
  beforeEach: () => {
    // applied の行にだけオファー/不採用ボタンが出る（遷移表の UI 写し）を確認できる
    fetchApplications.mockResolvedValue({ ok: true, data: applications })
  },
}

export const Empty: Story = {
  name: '空',
  beforeEach: () => {
    fetchApplications.mockResolvedValue({ ok: true, data: [] })
  },
}

export const Error: Story = {
  name: 'エラー',
  beforeEach: () => {
    fetchApplications.mockResolvedValue({
      ok: false,
      error: '案件が見つかりません',
    })
  },
}
