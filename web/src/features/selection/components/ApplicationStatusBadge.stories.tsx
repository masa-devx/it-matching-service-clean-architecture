import type { Meta, StoryObj } from '@storybook/nextjs-vite'
import { TsunaguWorksApplicationStatus } from '@repo/api-client/company/generated/models'

import { ApplicationStatusBadge } from './ApplicationStatusBadge'

// company 視点のラベル（applied = 「選考待ち」= 自分が動く番）。
// talent 視点（features/application 側）と並べて見ると、同じ状態のラベルと色が反転する
const meta = {
  title: 'Parts/応募ステータスバッジ（company視点）',
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
