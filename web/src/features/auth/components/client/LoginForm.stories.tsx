import type { Meta, StoryObj } from '@storybook/nextjs-vite'
import { expect } from 'storybook/test'

import { loginCompanyAction } from '../../actions/company.mock'
import { LoginForm } from './LoginForm'

// フォームの状態網羅: 初期 / 入力エラー / 送信中 / サーバーエラー。
// action はモック（main.ts の差し替え）なので、成功・失敗・応答なしを story ごとに再現できる
const meta = {
  title: 'Features/認証/ログインフォーム',
  component: LoginForm,
  args: { role: 'company' },
} satisfies Meta<typeof LoginForm>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = { name: '初期表示' }

export const ValidationError: Story = {
  name: '入力エラー',
  play: async ({ canvas, userEvent }) => {
    // 空のまま送信 → Zod が止める（サーバーには到達しない）
    await userEvent.click(canvas.getByRole('button', { name: 'ログイン' }))
    await expect((await canvas.findAllByRole('alert')).length).toBeGreaterThan(
      0,
    )
    await expect(loginCompanyAction).not.toHaveBeenCalled()
  },
}

export const Submitting: Story = {
  name: '送信中',
  beforeEach: () => {
    // 永遠に解決しない Promise = 送信中の状態を固定する
    loginCompanyAction.mockImplementation(() => new Promise(() => {}))
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.type(
      canvas.getByLabelText('メールアドレス'),
      'company1@example.com',
    )
    await userEvent.type(canvas.getByLabelText('パスワード'), 'password123')
    await userEvent.click(canvas.getByRole('button', { name: 'ログイン' }))
    await expect(
      await canvas.findByRole('button', { name: 'ログイン中…' }),
    ).toBeDisabled()
  },
}

export const ServerError: Story = {
  name: 'サーバーエラー',
  beforeEach: () => {
    loginCompanyAction.mockResolvedValue({
      error: 'メールアドレスまたはパスワードが正しくありません',
    })
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.type(
      canvas.getByLabelText('メールアドレス'),
      'company1@example.com',
    )
    await userEvent.type(canvas.getByLabelText('パスワード'), 'wrongpass1')
    await userEvent.click(canvas.getByRole('button', { name: 'ログイン' }))
    await expect(
      await canvas.findByText(
        'メールアドレスまたはパスワードが正しくありません',
      ),
    ).toBeInTheDocument()
  },
}
