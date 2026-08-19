import path from 'path'

import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  // ワークスペース内の TS パッケージ（生成クライアント）を Next.js 側でコンパイルする
  transpilePackages: ['@repo/api-client'],

  // 本番デプロイ用: 実行に必要な最小ファイル群を .next/standalone に自己完結で出力する
  // （node_modules 全体を持ち込まない＝イメージ削減・ADR-0011）
  output: 'standalone',
  // モノレポではトレースの基準をリポジトリルートに明示する
  // （ロックファイル位置からの推測に頼らない）
  outputFileTracingRoot: path.join(__dirname, '..'),
}

export default nextConfig
