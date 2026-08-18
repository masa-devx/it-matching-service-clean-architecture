import type { Meta, StoryObj } from '@storybook/nextjs-vite'
import type { TsunaguWorksProject } from '@repo/api-client/talent/generated/models'
import { expect } from 'storybook/test'

import { fetchProjects } from '../../actions/read.mock'
import { ProjectSearchList } from './ProjectSearchList'

// ページ級 story: /talent/projects の実体（無限リスト）。
// 「追加読み込み中」は無限リスト固有の本物のローディング状態（初回はルートの loading.tsx が担当）
const meta = {
  title: 'Pages/案件を探す',
  component: ProjectSearchList,
  args: { filters: { skills: [] } },
} satisfies Meta<typeof ProjectSearchList>

export default meta
type Story = StoryObj<typeof meta>

const page = (offset: number): TsunaguWorksProject[] =>
  [0, 1, 2].map((i) => ({
    id: offset - i,
    title: `デモ案件 #${offset - i}`,
    description: 'シード相当のダミー案件です。',
    hourly_rate_min: 3000 + i * 500,
    hourly_rate_max: 5000 + i * 500,
    hours_per_week: 10 + i * 5,
    remote_ok: i % 2 === 0,
    required_skills: ['Go', 'PostgreSQL'],
    status: 'published',
    created_at: '2026-08-15T09:00:00Z',
  }))

export const WithData: Story = {
  name: 'データあり（次ページあり）',
  beforeEach: () => {
    fetchProjects.mockResolvedValue({
      ok: true,
      data: { projects: page(42), next_cursor: 39 },
    })
  },
}

export const LastPage: Story = {
  name: 'データあり（最終ページ・もっと見る消滅）',
  beforeEach: () => {
    fetchProjects.mockResolvedValue({
      ok: true,
      data: { projects: page(3), next_cursor: null },
    })
  },
}

export const FetchingNextPage: Story = {
  name: '追加読み込み中',
  beforeEach: () => {
    // 1ページ目は即返し、2ページ目は永遠に解決しない = 「読み込み中…」を固定
    fetchProjects
      .mockResolvedValueOnce({
        ok: true,
        data: { projects: page(42), next_cursor: 39 },
      })
      .mockImplementation(() => new Promise(() => {}))
  },
  play: async ({ canvas, userEvent }) => {
    await userEvent.click(
      await canvas.findByRole('button', { name: 'もっと見る' }),
    )
    await expect(
      await canvas.findByRole('button', { name: '読み込み中…' }),
    ).toBeDisabled()
  },
}

export const Empty: Story = {
  name: '空（条件に合う案件なし）',
  beforeEach: () => {
    fetchProjects.mockResolvedValue({
      ok: true,
      data: { projects: [], next_cursor: null },
    })
  },
}

export const Error: Story = {
  name: 'エラー',
  beforeEach: () => {
    fetchProjects.mockResolvedValue({
      ok: false,
      error: 'サーバーに接続できませんでした',
    })
  },
}
