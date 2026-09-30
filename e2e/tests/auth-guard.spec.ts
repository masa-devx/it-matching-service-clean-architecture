import { expect, test } from '@playwright/test'

// 目的: ルートグループの layout.tsx によるガード（Cookie なし → Go の me が null → リダイレクト）が
// ブラウザから見て効いていることを固定する。壊れると未ログインで保護画面の枠が表示される。
// 観点: company / talent の両系統。ロール越境（talent で /company/*）は後続 Issue で追加する
const cases = [
  {
    role: 'company',
    protectedPath: '/company/dashboard',
    loginPath: /\/company\/login$/,
    heading: '企業ログイン',
  },
  {
    role: 'talent',
    protectedPath: '/talent/dashboard',
    loginPath: /\/talent\/login$/,
    heading: '人材ログイン',
  },
] as const

for (const c of cases) {
  test(`未ログインで ${c.protectedPath} を開くと ${c.role} のログイン画面へ送られる`, async ({
    page,
  }) => {
    await page.goto(c.protectedPath)

    await expect(page).toHaveURL(c.loginPath)
    await expect(page.getByRole('heading', { name: c.heading })).toBeVisible()
  })
}
