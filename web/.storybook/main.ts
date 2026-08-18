import type { StorybookConfig } from '@storybook/nextjs-vite'

import { dirname, resolve } from 'path'

import { fileURLToPath } from 'url'

/**
 * This function is used to resolve the absolute path of a package.
 * It is needed in projects that use Yarn PnP or are set up within a monorepo.
 */
function getAbsolutePath(value: string) {
  return dirname(fileURLToPath(import.meta.resolve(`${value}/package.json`)))
}

const configDir = dirname(fileURLToPath(import.meta.url))

// Server Action（'use server'）のモジュールは server-only に繋がるため、
// ブラウザだけの Storybook では読み込めない → 解決結果を .mock.ts に差し替える。
// キー = 本物のパス末尾・値 = モックファイル（stories からは .mock を直接 import して挙動を制御する）
const serverActionMocks: Record<string, string> = {
  'src/features/auth/actions/company.ts': resolve(
    configDir,
    '../src/features/auth/actions/company.mock.ts',
  ),
  'src/features/auth/actions/talent.ts': resolve(
    configDir,
    '../src/features/auth/actions/talent.mock.ts',
  ),
  'src/features/application/actions/create.ts': resolve(
    configDir,
    '../src/features/application/actions/create.mock.ts',
  ),
}

const config: StorybookConfig = {
  stories: ['../src/**/*.stories.@(js|jsx|mjs|ts|tsx)'],
  addons: [getAbsolutePath('@storybook/addon-a11y')],
  framework: getAbsolutePath('@storybook/nextjs-vite'),
  viteFinal: async (viteConfig) => {
    viteConfig.plugins ??= []
    viteConfig.plugins.push({
      name: 'mock-server-actions',
      enforce: 'pre',
      async resolveId(source, importer, options) {
        const resolved = await this.resolve(source, importer, options)
        if (!resolved) {
          return null
        }
        const id = resolved.id.replace(/\\/g, '/')
        const hit = Object.entries(serverActionMocks).find(([realPath]) =>
          id.endsWith(realPath),
        )
        return hit ? hit[1] : null
      },
    })
    return viteConfig
  },
}
export default config
