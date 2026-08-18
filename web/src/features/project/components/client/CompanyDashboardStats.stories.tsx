import type { Meta, StoryObj } from '@storybook/nextjs-vite'

import { fetchMyProjects } from '../../actions/read.mock'
import { CompanyDashboardStats } from './CompanyDashboardStats'

const projects = (statuses: string[]) =>
  statuses.map((status, i) => ({
    id: i + 1,
    title: `デモ案件 #${i + 1}`,
    description: 'ダミー',
    hourly_rate_min: null,
    hourly_rate_max: null,
    hours_per_week: 10,
    remote_ok: true,
    required_skills: [],
    status,
    created_at: '2026-08-15T09:00:00Z',
  }))

const meta = {
  title: 'Pages/ダッシュボード（company）/統計',
  component: CompanyDashboardStats,
} satisfies Meta<typeof CompanyDashboardStats>

export default meta
type Story = StoryObj<typeof meta>

export const WithData: Story = {
  name: 'データあり',
  beforeEach: () => {
    fetchMyProjects.mockResolvedValue({
      ok: true,
      data: projects([
        'published',
        'published',
        'published',
        'draft',
        'closed',
      ]),
    })
  },
}

export const Zero: Story = {
  name: '案件ゼロ',
  beforeEach: () => {
    fetchMyProjects.mockResolvedValue({ ok: true, data: [] })
  },
}
