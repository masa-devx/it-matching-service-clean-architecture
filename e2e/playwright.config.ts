import path from 'node:path'

import { defineConfig, devices } from '@playwright/test'

// ブラウザ E2E は「API 統合テストでは届かない継ぎ目」だけを検証する（ADR-0014）。
// 業務ルールはここに書かない（domain / usecase / apitest の担当）

const repoRoot = path.resolve(__dirname, '..')

// 開発サーバー（API :8082 / web :3001）と別ポートにする。既存サーバーを再利用すると
// 開発 DB に繋がったサーバーでテストが走り、結果が手元のデータに左右されるため
const API_PORT = 8092
const WEB_PORT = 3011
const WEB_URL = `http://localhost:${WEB_PORT}`

// E2E 専用 DB（make e2e-web-reset の DB 名ガードと同じ DB を指す）
const E2E_DATABASE_URL =
  process.env.E2E_WEB_DATABASE_URL ??
  'postgres://tsunagu:tsunagu@localhost:5435/tsunagu_e2e_web?sslmode=disable'

// テスト専用の署名鍵（本番とは無関係の公開値。apitest の jwtSecret と同じ扱い）
const JWT_SECRET = 'e2e-web-only-secret'

const isCI = !!process.env.CI

export default defineConfig({
  testDir: './tests',
  globalSetup: './global-setup.ts',

  // 最初は直列で安定させる。書き込むシナリオが「実行ごとに一意なデータ」を守れていると
  // 確認できてから並列度を上げる
  fullyParallel: false,
  workers: 1,
  forbidOnly: isCI,
  // ローカルは再試行しない（不安定さを再試行で隠さず、その場で気づく）
  retries: isCI ? 2 : 0,
  reporter: isCI ? [['list'], ['html', { open: 'never' }]] : 'list',

  use: {
    baseURL: WEB_URL,
    // 失敗したテストだけ trace（操作・DOM・ネットワークの記録）を残す
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },

  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],

  webServer: [
    {
      // API はサーバー起動時に DB へ Ping するため、DB の作成・マイグレーションを先に済ませる
      // （webServer と globalSetup の実行順に依存しない）
      command: 'make e2e-web-db-setup && cd api-server && go run ./cmd/server',
      cwd: repoRoot,
      url: `http://localhost:${API_PORT}/health`,
      env: {
        DATABASE_URL: E2E_DATABASE_URL,
        JWT_SECRET,
        PORT: String(API_PORT),
        // .env に設定があっても E2E ではトレースを送らない（godotenv は既存の環境変数を上書きしない）
        OTEL_EXPORTER_OTLP_ENDPOINT: '',
      },
      reuseExistingServer: false,
      timeout: 120_000,
      stdout: 'ignore',
      stderr: 'pipe',
    },
    {
      // 本番（Cloud Run・web/Dockerfile）と同じ standalone 出力で起動する。
      // dev モードはページを開いた時点でコンパイルするため、待ち時間が揺れて不安定になる。
      // ビルドは turbo の test:e2e が依存タスク（web#build）として先に済ませる
      command:
        'rm -rf .next/standalone/web/.next/static && cp -R .next/static .next/standalone/web/.next/static && node .next/standalone/web/server.js',
      cwd: path.join(repoRoot, 'web'),
      url: WEB_URL,
      env: {
        PORT: String(WEB_PORT),
        HOSTNAME: 'localhost',
        API_URL: `http://localhost:${API_PORT}`,
      },
      reuseExistingServer: false,
      timeout: 60_000,
      stdout: 'ignore',
      stderr: 'pipe',
    },
  ],
})
