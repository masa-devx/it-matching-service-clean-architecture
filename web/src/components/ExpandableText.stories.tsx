import type { Meta, StoryObj } from '@storybook/nextjs-vite'
import { expect } from 'storybook/test'

import { ExpandableText } from './ExpandableText'

const meta = {
  title: 'Parts/折りたたみテキスト',
  component: ExpandableText,
} satisfies Meta<typeof ExpandableText>

export default meta
type Story = StoryObj<typeof meta>

export const Short: Story = {
  name: '短文（トグルなし）',
  args: { text: 'バックエンドの経験を活かせると思い応募しました。' },
}

export const Long: Story = {
  name: '長文（折りたたみ + 展開の動作確認）',
  args: {
    text: 'Go と PostgreSQL を用いたバックエンド開発に5年従事してきました。\n直近ではマッチングサービスの API 設計・実装を担当し、状態機械を持つドメインの設計やシークページネーションの実装経験があります。\n副業として週15〜20時間の稼働が可能です。よろしくお願いいたします。',
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(canvas.getByRole('button', { name: 'もっと見る' }))
    await expect(
      await canvas.findByRole('button', { name: '閉じる' }),
    ).toBeInTheDocument()
  },
}
