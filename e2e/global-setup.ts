import { execSync } from 'node:child_process'
import path from 'node:path'

// テストの前に毎回、E2E 専用 DB を基準世界へ戻す（ADR-0014）。
// 実サーバーが COMMIT するため後始末では消せず、「前にまっさらにする」方式で
// 前回の実行の残り・途中失敗の影響を断つ。
// 接続先は E2E_WEB_DATABASE_URL（未設定なら cmd/e2ereset のローカル既定値）
export default function globalSetup() {
  execSync('make e2e-web-reset', {
    cwd: path.resolve(__dirname, '..'),
    stdio: 'inherit',
  })
}
