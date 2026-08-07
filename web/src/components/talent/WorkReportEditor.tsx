'use client'

import { useState } from 'react'
import type { WorkReport } from '@/lib/workReports'
import { WorkReportForm } from '@/components/talent/WorkReportForm'
import { Button } from '@/components/ui/button'

// 差し戻された報告を修正するための開閉。
//
// 常にフォームを開いておくと、報告が何件もある画面が入力欄だらけになるため、
// 「修正する」を押したときだけ開く。開閉の状態を持つのでクライアント側の部品にしている
export function WorkReportEditor({ report }: { report: WorkReport }) {
  const [editing, setEditing] = useState(false)

  if (!editing) {
    return (
      <Button
        variant="outline"
        className="h-11 self-start"
        onClick={() => setEditing(true)}
      >
        修正して再提出する
      </Button>
    )
  }

  return (
    <div className="flex flex-col gap-4 rounded-md border bg-background p-4">
      <WorkReportForm contractId={report.contract_id} report={report} />
      <Button
        variant="ghost"
        className="h-11 self-start"
        onClick={() => setEditing(false)}
      >
        キャンセル
      </Button>
    </div>
  )
}
