import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

// 各テスト後にDOMを破棄する（テスト間の汚染防止）
afterEach(() => {
  cleanup()
})
