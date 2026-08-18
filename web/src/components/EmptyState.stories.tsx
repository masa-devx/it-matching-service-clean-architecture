import type { Meta, StoryObj } from '@storybook/nextjs-vite'
import { Inbox, SearchX } from 'lucide-react'

import { Button } from '@/components/ui/button'

import { EmptyState } from './EmptyState'

const meta = {
  title: 'Parts/空状態',
  component: EmptyState,
  args: {
    icon: Inbox,
    title: 'まだ応募がありません',
  },
} satisfies Meta<typeof EmptyState>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const WithAction: Story = {
  name: '説明 + 次の行動つき',
  args: {
    icon: SearchX,
    title: '条件に合う案件が見つかりませんでした',
    description: '条件を減らして試してください。',
    children: <Button variant="outline">条件をクリア</Button>,
  },
}
