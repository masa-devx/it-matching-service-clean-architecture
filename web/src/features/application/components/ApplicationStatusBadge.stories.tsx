import type { Meta, StoryObj } from '@storybook/nextjs-vite'
import { TsunaguWorksApplicationStatus } from '@repo/api-client/talent/generated/models'

import { ApplicationStatusBadge } from './ApplicationStatusBadge'

// talent 視点のラベル（offered = 「オファーが届いています」= 自分が動く番）。
// company 視点（features/selection 側）との対比が状態機械の actor 列の可視化になっている
const meta = {
  title: 'Parts/応募ステータスバッジ（talent視点）',
  component: ApplicationStatusBadge,
} satisfies Meta<typeof ApplicationStatusBadge>

export default meta
type Story = StoryObj<typeof meta>

export const Applied: Story = { args: { status: 'applied' } }
export const Offered: Story = { args: { status: 'offered' } }
export const Accepted: Story = { args: { status: 'accepted' } }

export const AllStates: Story = {
  name: '全状態',
  args: { status: 'applied' },
  render: () => (
    <div className="flex flex-wrap gap-2">
      {Object.values(TsunaguWorksApplicationStatus).map((status) => (
        <ApplicationStatusBadge key={status} status={status} />
      ))}
    </div>
  ),
}
