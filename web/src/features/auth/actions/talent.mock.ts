import { fn } from 'storybook/test'

// Storybook 用のモック（company.mock.ts と同じ仕組み）
export const signupTalentAction = fn().mockName('signupTalentAction')
export const loginTalentAction = fn().mockName('loginTalentAction')
