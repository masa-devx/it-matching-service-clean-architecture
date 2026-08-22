#!/usr/bin/env bash
# HTTP 層のパフォーマンス計測（P-2・#115）。
# 前提: make dev-api で API が起動済み・make seed-perf 投入済み（perf ユーザーでログインする）
# 使い方: make perf-measure（結果は標準出力。docs/パフォーマンス.md へ転記する）
#
# hey は go run で実行するためインストール不要（版固定）。
# -n 300 -c 10: 300リクエストを並列10で。単発 curl では p99（たまに遅い）が測れない
set -euo pipefail

BASE="${BASE:-http://localhost:8082}"
HEY="go run github.com/rakyll/hey@v0.1.4"

# 本物のログインでトークン取得（perfseed が作った人材ユーザー・パスワードは固定値）
token=$(curl -sf -X POST "${BASE}/talent/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"perf-talent-1@example.com","password":"e2ePassword123"}' | jq -r '.token')
if [ -z "${token}" ] || [ "${token}" = "null" ]; then
  echo "ログインに失敗しました（make dev-api と make seed-perf を確認してください）" >&2
  exit 1
fi

measure() {
  local name="$1" url="$2"
  echo "== ${name} =="
  # Latency distribution の 50% / 99% 行と平均だけを抜き出す
  ${HEY} -n 300 -c 10 -H "Authorization: Bearer ${token}" "${url}" 2>/dev/null |
    grep -E "Average:|50% in|99% in"
  echo
}

echo "計測開始: ${BASE}（300リクエスト×並列10）"
echo

measure "素の公開一覧（1ページ目）"        "${BASE}/talent/projects"
measure "skills 中間帯（Rust,Terraform）"  "${BASE}/talent/projects?skills=Rust,Terraform"
measure "時給 >= 8000（該当0件・本命）"    "${BASE}/talent/projects?min_hourly_rate=8000"
measure "時給 >= 5000（ヒットあり）"       "${BASE}/talent/projects?min_hourly_rate=5000"
measure "自分の応募一覧"                   "${BASE}/talent/applications"
