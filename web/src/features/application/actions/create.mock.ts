import { fn } from 'storybook/test'

// Storybook 用のモック（auth 側と同じ仕組み）
export const createApplicationAction = fn().mockName('createApplicationAction')
