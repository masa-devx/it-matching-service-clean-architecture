'use client'

import { AlertCircle } from 'lucide-react'
import type { FieldErrors } from 'react-hook-form'

// フィールドが多いフォームでは、各項目の直下に出るエラーを見落とす。
// 上部に一覧を出し、クリックで該当フィールドへ飛べるようにする
export function FormErrorSummary({
  errors,
  labels,
}: {
  errors: FieldErrors
  // フィールド名 → 画面上のラベル（「title」ではなく「案件タイトル」と見せる）
  labels: Record<string, string>
}) {
  const entries = Object.entries(errors).filter(([, error]) => error?.message)
  if (entries.length === 0) {
    return null
  }

  return (
    // role="alert" で出現時に読み上げられる（送信直後に気づける）
    <div
      role="alert"
      className="flex flex-col gap-2 rounded-lg border border-destructive/40 bg-destructive/5 p-4"
    >
      <p className="flex items-center gap-2 text-sm font-medium text-destructive">
        <AlertCircle className="size-4" aria-hidden="true" />
        {entries.length}件の入力を確認してください
      </p>
      <ul className="flex flex-col gap-1 text-sm">
        {entries.map(([name, error]) => (
          <li key={name}>
            {/* href="#id" で該当フィールドへ移動する（フィールドの id と揃える） */}
            <a href={`#${name}`} className="text-destructive underline">
              {labels[name] ?? name}: {String(error?.message)}
            </a>
          </li>
        ))}
      </ul>
    </div>
  )
}
