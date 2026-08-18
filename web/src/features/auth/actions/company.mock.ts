import { fn } from 'storybook/test'

// Storybook 用のモック（.storybook/main.ts の resolveId プラグインが本物と差し替える）。
// 本物は 'use server' で external/handler（server-only）に繋がるため、
// ブラウザだけの Storybook では読み込めない。fn() は呼び出しを記録し、
// story ごとに mockResolvedValue 等で挙動を差し替えられる
export const signupCompanyAction = fn().mockName('signupCompanyAction')
export const loginCompanyAction = fn().mockName('loginCompanyAction')
