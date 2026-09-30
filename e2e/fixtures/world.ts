// 基準世界（api-server/test/e2efixture）のうち、ブラウザ E2E が参照する値。
// 一次情報は ids.go / seed.go。基準世界を変えたらここも合わせる
// （件数・ID の網羅的な検証は Go 側の TestLoad が担う）

export const PASSWORD = 'e2ePassword123'

export const companyA = {
  email: 'company-a@example.com',
  name: '株式会社エー',
} as const

export const talentA = {
  email: 'talent-a@example.com',
  displayName: '人材エー',
} as const
