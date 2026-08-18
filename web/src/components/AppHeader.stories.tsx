import type { Meta, StoryObj } from '@storybook/nextjs-vite'

import { Button } from '@/components/ui/button'

import { AppHeader } from './AppHeader'

// children スロット設計のおかげで、ログアウト（features/auth・Server Action）を
// モックせずに素のボタンで代替できる＝「表示と依存の分離」が story の書きやすさに直結する
const meta = {
  title: 'Parts/アプリヘッダー',
  component: AppHeader,
  parameters: {
    layout: 'fullscreen',
    nextjs: { appDirectory: true },
  },
  args: {
    role: 'company',
    displayName: '株式会社テックワークス',
    children: (
      <Button variant="ghost" size="sm">
        ログアウト
      </Button>
    ),
  },
} satisfies Meta<typeof AppHeader>

export default meta
type Story = StoryObj<typeof meta>

export const Company: Story = {
  name: 'company（ダッシュボード表示中）',
  parameters: {
    nextjs: {
      appDirectory: true,
      navigation: { pathname: '/company/dashboard' },
    },
  },
}

export const Talent: Story = {
  name: 'talent（案件を探す表示中）',
  args: { role: 'talent', displayName: '山田太郎' },
  parameters: {
    nextjs: {
      appDirectory: true,
      navigation: { pathname: '/talent/projects' },
    },
  },
}

export const ActiveOnChildPage: Story = {
  name: '現在地強調（編集ページでも案件管理が光る）',
  parameters: {
    nextjs: {
      appDirectory: true,
      navigation: { pathname: '/company/projects/5/edit' },
    },
  },
}

export const Mobile: Story = {
  name: 'モバイル（ハンバーガー表示）',
  globals: { viewport: { value: 'mobile1' } },
  parameters: {
    nextjs: {
      appDirectory: true,
      navigation: { pathname: '/company/dashboard' },
    },
  },
}
