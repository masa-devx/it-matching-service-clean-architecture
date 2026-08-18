import type { Preview } from '@storybook/nextjs-vite'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

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
  decorators: [
    // useQuery を使う部品のための QueryClient。story ごとに新規作成する
    // （キャッシュが story 間で混ざらない）。retry: false はエラー系 story の即時表示に必須
    // （既定の3回リトライだとエラーが出るまで数秒待たされる）
    (Story) => {
      const queryClient = new QueryClient({
        defaultOptions: { queries: { retry: false } },
      })
      return (
        <QueryClientProvider client={queryClient}>
          <Story />
        </QueryClientProvider>
      )
    },
  ],
}

export default preview
