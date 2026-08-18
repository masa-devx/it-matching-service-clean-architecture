import type { Preview } from '@storybook/nextjs-vite'

// Tailwind v4 のトークン・ベーススタイルを読み込む（無いと全部品が素の HTML になる）
import '../src/app/globals.css'

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
}

export default preview
