import type { Meta, StoryObj } from '@storybook/nextjs-vite'
import type { TsunaguWorksApplication } from '@repo/api-client/talent/generated/models'

import { fetchMyApplications } from '../../actions/read.mock'
import { ApplicationList } from './ApplicationList'

// ページ級 story: /talent/applications の実体。
// データありでは全6状態を並べる＝状態ごとの操作ボタンの違い
// （applied=取り下げのみ・offered=承諾/辞退/取り下げ・終端=なし）が1画面で見える
const meta = {
  title: 'Pages/応募一覧（talent）',
  component: ApplicationList,
} satisfies Meta<typeof ApplicationList>

export default meta
type Story = StoryObj<typeof meta>

const applications: TsunaguWorksApplication[] = [
  {
    id: 6,
    project_id: 6,
    project_title: 'Go バックエンドの API 開発',
    status: 'offered',
    message: 'バックエンドの経験を活かせると思い応募しました。',
    created_at: '2026-08-16T09:00:00Z',
  },
  {
    id: 5,
    project_id: 5,
    project_title: '管理画面のフロントエンド改修',
    status: 'applied',
    message: '',
    created_at: '2026-08-15T09:00:00Z',
  },
  {
    id: 4,
    project_id: 4,
    project_title: 'インフラの IaC 化',
    status: 'accepted',
    message: 'Terraform の実務経験があります。',
    created_at: '2026-08-10T09:00:00Z',
  },
  {
    id: 3,
    project_id: 3,
    project_title: 'モバイルアプリの API 連携',
    status: 'rejected',
    message: '',
    created_at: '2026-08-05T09:00:00Z',
  },
  {
    id: 2,
    project_id: 2,
    project_title: 'データ基盤の構築',
    status: 'withdrawn',
    message: '',
    created_at: '2026-08-03T09:00:00Z',
  },
  {
    id: 1,
    project_id: 1,
    project_title: '社内ツールの保守',
    status: 'declined',
    message: '',
    created_at: '2026-08-01T09:00:00Z',
  },
]

export const WithData: Story = {
  name: 'データあり（全6状態）',
  beforeEach: () => {
    fetchMyApplications.mockResolvedValue({ ok: true, data: applications })
  },
}

export const Empty: Story = {
  name: '空',
  beforeEach: () => {
    fetchMyApplications.mockResolvedValue({ ok: true, data: [] })
  },
}

export const Error: Story = {
  name: 'エラー',
  beforeEach: () => {
    fetchMyApplications.mockResolvedValue({
      ok: false,
      error: 'サーバーに接続できませんでした',
    })
  },
}
