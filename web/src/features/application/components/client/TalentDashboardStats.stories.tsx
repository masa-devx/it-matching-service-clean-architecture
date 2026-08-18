import type { Meta, StoryObj } from '@storybook/nextjs-vite'

import { fetchMyApplications } from '../../actions/read.mock'
import { TalentDashboardStats } from './TalentDashboardStats'

const apps = (statuses: string[]) =>
  statuses.map((status, i) => ({
    id: i + 1,
    project_id: i + 1,
    project_title: `デモ案件 #${i + 1}`,
    status,
    message: '',
    created_at: '2026-08-15T09:00:00Z',
  }))

const meta = {
  title: 'Pages/ダッシュボード（talent）/統計',
  component: TalentDashboardStats,
} satisfies Meta<typeof TalentDashboardStats>

export default meta
type Story = StoryObj<typeof meta>

export const WithOffer: Story = {
  name: 'オファーあり（強調）',
  beforeEach: () => {
    fetchMyApplications.mockResolvedValue({
      ok: true,
      data: apps(['offered', 'offered', 'applied', 'accepted', 'rejected']),
    })
  },
}

export const NoOffer: Story = {
  name: 'オファーなし',
  beforeEach: () => {
    fetchMyApplications.mockResolvedValue({
      ok: true,
      data: apps(['applied', 'applied', 'withdrawn']),
    })
  },
}

export const Zero: Story = {
  name: '応募ゼロ',
  beforeEach: () => {
    fetchMyApplications.mockResolvedValue({ ok: true, data: [] })
  },
}
