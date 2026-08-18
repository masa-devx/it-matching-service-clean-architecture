import type { Meta, StoryObj } from '@storybook/nextjs-vite'
import type { TsunaguWorksProject } from '@repo/api-client/company/generated/models'

import { fetchMyProjects } from '../../actions/read.mock'
import { ProjectList } from './ProjectList'

// ページ級 story: /company/projects の実体。データあり / 空 / エラーを切り替えて確認する
// （ローディングはルート境界の loading.tsx が担う設計のため、リスト部品には存在しない）
const meta = {
  title: 'Pages/案件管理',
  component: ProjectList,
} satisfies Meta<typeof ProjectList>

export default meta
type Story = StoryObj<typeof meta>

const projects: TsunaguWorksProject[] = [
  {
    id: 3,
    title: 'Go バックエンドの API 開発',
    description: 'マッチングサービスの API 開発をお願いします。',
    hourly_rate_min: 4000,
    hourly_rate_max: 6000,
    hours_per_week: 20,
    remote_ok: true,
    required_skills: ['Go', 'PostgreSQL'],
    status: 'published',
    created_at: '2026-08-15T09:00:00Z',
  },
  {
    id: 2,
    title: '管理画面のフロントエンド改修',
    description: 'Next.js の管理画面の改修です。',
    hourly_rate_min: null,
    hourly_rate_max: null,
    hours_per_week: 10,
    remote_ok: false,
    required_skills: ['TypeScript', 'React'],
    status: 'draft',
    created_at: '2026-08-10T09:00:00Z',
  },
  {
    id: 1,
    title: 'インフラの IaC 化',
    description: 'Terraform への移行を支援してください。',
    hourly_rate_min: 5000,
    hourly_rate_max: 7000,
    hours_per_week: 15,
    remote_ok: true,
    required_skills: ['AWS', 'Terraform'],
    status: 'closed',
    created_at: '2026-08-01T09:00:00Z',
  },
]

export const WithData: Story = {
  name: 'データあり',
  beforeEach: () => {
    // 3状態を混ぜる＝状態ごとの操作ボタンの違い（公開/非公開/終了/再公開）が1画面で見える
    fetchMyProjects.mockResolvedValue({ ok: true, data: projects })
  },
}

export const Empty: Story = {
  name: '空',
  beforeEach: () => {
    fetchMyProjects.mockResolvedValue({ ok: true, data: [] })
  },
}

export const Error: Story = {
  name: 'エラー',
  beforeEach: () => {
    fetchMyProjects.mockResolvedValue({
      ok: false,
      error: 'サーバーに接続できませんでした',
    })
  },
}
