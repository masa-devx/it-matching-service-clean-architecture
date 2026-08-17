import { defineConfig, globalIgnores } from 'eslint/config'
import nextVitals from 'eslint-config-next/core-web-vitals'
import nextTs from 'eslint-config-next/typescript'
import prettier from 'eslint-config-prettier/flat'

const eslintConfig = defineConfig([
  ...nextVitals,
  ...nextTs,
  // 依存方向の強制（設計プラン§4・Phase 2 の features / external 構成を先取りしたガードレール）。
  // import の規約: 境界を越えるときはエイリアス（@/）・同一領域内は相対パス、と書き分ける。
  // この規約の上で「エイリアスでの越境」だけを禁止すれば、境界ルールが標準ルールで表現できる
  {
    rules: {
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['@/external/client', '@/external/client/*'],
              message:
                'external/client（Go API 通信）は直接 import せず、external/handler 経由で使う（隣接する handler からは相対パスで import する）',
            },
            {
              group: ['@/features/*'],
              message:
                'feature 間の直接依存は禁止。同一 feature 内は相対パスで import する（共有したいものは components / hooks / lib へ昇格させる）',
            },
          ],
        },
      ],
    },
  },
  // 整形はPrettierに任せるため、ESLint側の整形系ルールを無効化（必ず最後に置く）
  prettier,
  // Override default ignores of eslint-config-next.
  globalIgnores([
    // Default ignores of eslint-config-next:
    '.next/**',
    'out/**',
    'build/**',
    'next-env.d.ts',
  ]),
])

export default eslintConfig
