import type { NextConfig } from 'next'

const nextConfig: NextConfig = {
  // ワークスペース内の TS パッケージ（生成クライアント）を Next.js 側でコンパイルする
  transpilePackages: ['@repo/api-client'],
}

export default nextConfig
