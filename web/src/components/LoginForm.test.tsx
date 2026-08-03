import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { LoginForm } from './LoginForm'

// vi.mock はファイル先頭に巻き上げられるため、モック関数は vi.hoisted で先に作る
const { loginMock, pushMock } = vi.hoisted(() => ({
  loginMock: vi.fn(),
  pushMock: vi.fn(),
}))

// 通信は lib/authClient に集約されているので、この1関数を差し替えるだけでテストできる
vi.mock('@/lib/authClient', () => ({
  login: loginMock,
}))

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: pushMock, refresh: vi.fn() }),
}))

async function fillAndSubmit(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText('メールアドレス'), 'user@example.com')
  await user.type(screen.getByLabelText('パスワード'), 'password123')
  await user.click(screen.getByRole('button', { name: 'ログイン' }))
}

describe('LoginForm', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('ログイン失敗時にエラーメッセージが表示され、遷移しない', async () => {
    loginMock.mockResolvedValue({
      ok: false,
      error: 'メールアドレスまたはパスワードが正しくありません',
    })
    const user = userEvent.setup()
    render(<LoginForm />)

    await fillAndSubmit(user)

    // role="alert" の要素として出ること自体も仕様（スクリーンリーダー通知）
    expect(await screen.findByRole('alert')).toHaveTextContent(
      'メールアドレスまたはパスワードが正しくありません',
    )
    expect(pushMock).not.toHaveBeenCalled()
    // 失敗後は再送信できるようボタンが復活している
    expect(screen.getByRole('button', { name: 'ログイン' })).toBeEnabled()
  })

  it('ログイン成功時に / へ遷移し、ボタンは無効のまま', async () => {
    loginMock.mockResolvedValue({ ok: true })
    const user = userEvent.setup()
    render(<LoginForm />)

    await fillAndSubmit(user)

    await vi.waitFor(() => {
      expect(pushMock).toHaveBeenCalledWith('/')
    })
    expect(loginMock).toHaveBeenCalledWith({
      email: 'user@example.com',
      password: 'password123',
    })
    // 遷移完了までボタンを無効に保つ仕様（二重送信防止）
    expect(screen.getByRole('button', { name: 'ログイン中…' })).toBeDisabled()
  })
})
