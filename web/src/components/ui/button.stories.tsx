import type { Meta, StoryObj } from '@storybook/nextjs-vite'

import { Button } from './button'

// shadcn 生成部品のカタログ（生成物本体は編集しない。stories は併設ファイル）
const meta = {
  title: 'UI/Button',
  component: Button,
  args: { children: 'ボタン' },
} satisfies Meta<typeof Button>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const AllVariants: Story = {
  name: '全 variant × 状態',
  render: () => (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-2">
        {(
          [
            'default',
            'outline',
            'secondary',
            'ghost',
            'destructive',
            'link',
          ] as const
        ).map((variant) => (
          <Button key={variant} variant={variant}>
            {variant}
          </Button>
        ))}
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <Button disabled>disabled</Button>
        <Button size="sm">sm</Button>
        <Button size="lg">lg</Button>
      </div>
    </div>
  ),
}
