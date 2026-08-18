import type { Meta, StoryObj } from '@storybook/nextjs-vite'

import { StatCard } from './StatCard'

const meta = {
  title: 'Parts/統計カード',
  component: StatCard,
  args: { label: '掲載中の案件', value: 12, href: '#' },
} satisfies Meta<typeof StatCard>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const Highlight: Story = {
  name: '強調（今やるべきことがある）',
  args: { label: 'オファーが届いています', value: 2, highlight: true },
}

export const Zero: Story = {
  name: 'ゼロ',
  args: { value: 0 },
}
