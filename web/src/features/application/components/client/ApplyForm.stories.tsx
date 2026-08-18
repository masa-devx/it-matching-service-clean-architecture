import type { Meta, StoryObj } from '@storybook/nextjs-vite'
import { expect } from 'storybook/test'

import { createApplicationAction } from '../../actions/create.mock'
import { ApplyForm } from './ApplyForm'

const meta = {
  title: 'Features/応募/応募フォーム',
  component: ApplyForm,
  args: { projectId: 1 },
} satisfies Meta<typeof ApplyForm>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = { name: '初期表示' }

export const Submitting: Story = {
  name: '送信中',
  beforeEach: () => {
    createApplicationAction.mockImplementation(() => new Promise(() => {}))
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(
      canvas.getByRole('button', { name: 'この案件に応募する' }),
    )
    await expect(
      await canvas.findByRole('button', { name: '応募中…' }),
    ).toBeDisabled()
  },
}

// interaction test の見本: 入力 → 送信 → 二重応募 409 の表示までを自動再生する「動くドキュメント」
export const ServerError: Story = {
  name: 'サーバーエラー（二重応募）',
  beforeEach: () => {
    createApplicationAction.mockResolvedValue({
      error: 'この案件にはすでに応募しています',
    })
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.type(
      canvas.getByLabelText('志望動機（任意）'),
      'Go と PostgreSQL の経験が活かせると思い応募しました。',
    )
    await userEvent.click(
      canvas.getByRole('button', { name: 'この案件に応募する' }),
    )
    await expect(
      await canvas.findByText('この案件にはすでに応募しています'),
    ).toBeInTheDocument()
    await expect(createApplicationAction).toHaveBeenCalledWith(
      1,
      'Go と PostgreSQL の経験が活かせると思い応募しました。',
    )
  },
}
