'use client'

import { useState } from 'react'

import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

// 長文の折りたたみ表示（志望動機など）。
// 「実際に何行になるか」の判定は DOM 測定が必要でリサイズ追従も複雑になるため、
// 文字数と改行数のしきい値で近似する（誤差は「短いのにボタンが出る」方向にしか倒れない＝安全）。
// shadcn の Collapsible は不採用: line-clamp + トグルで足りる（部品を増やさない）
export function ExpandableText({ text }: { text: string }) {
  const [expanded, setExpanded] = useState(false)

  const isLong = text.length > 100 || text.split('\n').length > 2
  if (!isLong) {
    return <p className="text-sm whitespace-pre-wrap">{text}</p>
  }

  return (
    <div className="space-y-1">
      <p
        className={cn(
          'text-sm whitespace-pre-wrap',
          !expanded && 'line-clamp-2',
        )}
      >
        {text}
      </p>
      <Button
        type="button"
        variant="link"
        size="xs"
        className="px-0"
        onClick={() => setExpanded((v) => !v)}
      >
        {expanded ? '閉じる' : 'もっと見る'}
      </Button>
    </div>
  )
}
