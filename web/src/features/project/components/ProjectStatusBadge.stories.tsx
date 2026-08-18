import type { Meta, StoryObj } from '@storybook/nextjs-vite'
import { TsunaguWorksProjectStatus } from '@repo/api-client/company/generated/models'

import { ProjectStatusBadge } from './ProjectStatusBadge'

// 状態網羅 story の型: 個別の状態 + 全状態を1画面に並べる一覧を必ず置く
// （デザイン漏れ・ラベルの不整合に目視で気づける）
const meta = {
  title: 'Parts/案件ステータスバッジ',
  component: ProjectStatusBadge,
} satisfies Meta<typeof ProjectStatusBadge>

export default meta
type Story = StoryObj<typeof meta>

export const Draft: Story = { args: { status: 'draft' } }
export const Published: Story = { args: { status: 'published' } }
export const Closed: Story = { args: { status: 'closed' } }

export const AllStates: Story = {
  name: '全状態',
  args: { status: 'draft' },
  render: () => (
    <div className="flex gap-2">
      {Object.values(TsunaguWorksProjectStatus).map((status) => (
        <ProjectStatusBadge key={status} status={status} />
      ))}
    </div>
  ),
}
