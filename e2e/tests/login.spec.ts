import { expect, test } from '@playwright/test'

import { companyA, PASSWORD } from '../fixtures/world'

// 目的: ログインの継ぎ目（ブラウザ → Server Action → Go → httpOnly Cookie → RSC が Cookie を Go へ転送）を
// 通しで固定する。壊れるとログインできても次の画面で未ログイン扱いになる。
// 観点: ログイン後に基準世界の企業名が表示される（E2E 専用 DB の世界が読み込まれている証拠を兼ねる）／
// トークン Cookie が httpOnly（ブラウザ JS から読めない）／リロードしてもログインが続く
test('企業がログインするとダッシュボードが表示され、リロードしてもログインが続く', async ({
  page,
  context,
}) => {
  await page.goto('/company/login')
  await page.getByLabel('メールアドレス').fill(companyA.email)
  await page.getByLabel('パスワード').fill(PASSWORD)
  await page.getByRole('button', { name: 'ログイン' }).click()

  await expect(page).toHaveURL(/\/company\/dashboard$/)
  await expect(
    page.getByRole('heading', { name: '企業ダッシュボード' }),
  ).toBeVisible()
  await expect(page.getByRole('main').getByText(companyA.name)).toBeVisible()

  const token = (await context.cookies()).find((c) => c.name === 'token')
  expect(token?.httpOnly).toBe(true)

  await page.reload()
  await expect(page).toHaveURL(/\/company\/dashboard$/)
  await expect(page.getByRole('main').getByText(companyA.name)).toBeVisible()
})
