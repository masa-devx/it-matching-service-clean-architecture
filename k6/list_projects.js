// 案件一覧APIの負荷試験（実行: make perf-load）
//
// 目的: 「1リクエストの速さ」ではなく「同時アクセス時の応答分布」を測る。
// 平均は外れ値に引っ張られないため実態を隠す。実務では p95 / p99 を見る
// （100人に1人が3秒待っていても平均は200msに見えることがある）。
import http from 'k6/http'
import { check, sleep } from 'k6'

const BASE_URL = __ENV.BASE_URL || 'http://host.docker.internal:8081'
const EMAIL = __ENV.EMAIL || 'talent1@example.com'
const PASSWORD = __ENV.PASSWORD || 'password123'

export const options = {
  // 段階的に負荷を上げる（いきなり最大にすると起動直後の遅さを拾ってしまう）
  stages: [
    { duration: '10s', target: 10 }, // 10秒かけて同時10ユーザーへ
    { duration: '20s', target: 10 }, // 20秒維持（この区間が本計測）
    { duration: '5s', target: 0 },   // 収束
  ],
  thresholds: {
    // 満たせなければ k6 が失敗終了する = CIで性能の退行を検知できる
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
}

// setup は全VU共通で1度だけ実行される。トークン取得を各リクエストで繰り返さない
export function setup() {
  const res = http.post(
    `${BASE_URL}/login`,
    JSON.stringify({ email: EMAIL, password: PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } },
  )
  check(res, { 'login succeeded': (r) => r.status === 200 })
  return { token: res.json('token') }
}

export default function (data) {
  const params = {
    headers: { Authorization: `Bearer ${data.token}` },
  }

  // 実際の利用に近い比率で3種類のリクエストを混ぜる
  const scenarios = [
    `${BASE_URL}/projects?limit=20`,                          // 一覧
    `${BASE_URL}/projects?skills=Go&limit=20`,                // スキル検索
    `${BASE_URL}/projects?skills=Go&remote_ok=true&rate_min=8000&limit=20`, // 複合条件
  ]
  const url = scenarios[Math.floor(Math.random() * scenarios.length)]

  const res = http.get(url, params)
  check(res, {
    'status is 200': (r) => r.status === 200,
    'has projects': (r) => r.json('projects') !== undefined,
  })

  sleep(1) // 実ユーザーの操作間隔を模す（無しだと非現実的な連打になる）
}
