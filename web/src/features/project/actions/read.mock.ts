import { fn } from 'storybook/test'

// Storybook 用のモック（.storybook/main.ts が本物と差し替える）
export const fetchMyProjects = fn().mockName('fetchMyProjects')
export const fetchMyProject = fn().mockName('fetchMyProject')
