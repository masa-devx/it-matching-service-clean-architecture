import type { Meta, StoryObj } from '@storybook/nextjs-vite'

import { SkillBadges } from './SkillBadges'

const meta = {
  title: 'Parts/スキルバッジ',
  component: SkillBadges,
} satisfies Meta<typeof SkillBadges>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  args: { skills: ['Go', 'PostgreSQL', 'TypeScript'] },
}

export const Many: Story = {
  name: '多数（折り返し）',
  args: {
    skills: [
      'Go',
      'PostgreSQL',
      'TypeScript',
      'React',
      'Next.js',
      'AWS',
      'Docker',
      'Terraform',
      'Kubernetes',
    ],
  },
}

export const Empty: Story = {
  name: '空（何も描画しない）',
  args: { skills: [] },
}
