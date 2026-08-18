'use client'

import { useState } from 'react'

import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

import { createApplicationAction } from '../../actions/create'

// 応募フォーム。項目は志望動機（任意）1つだけなので RHF + Zod は使わない
// （必須項目も相関検証も無い入力に重装備は過剰・検索フォーム #46 と同じ判断）。
// 成功時は action 内の redirect で応募一覧へ遷移する
export function ApplyForm({ projectId }: { projectId: number }) {
  const [serverError, setServerError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const onSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setServerError(null)
    setSubmitting(true)
    const message = String(
      new FormData(e.currentTarget).get('message') ?? '',
    ).trim()

    const result = await createApplicationAction(projectId, message)
    if (result?.error) {
      setServerError(result.error)
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={onSubmit} className="space-y-3 rounded-lg border p-4">
      <div className="space-y-1.5">
        <Label htmlFor="message">志望動機（任意）</Label>
        <Textarea
          id="message"
          name="message"
          rows={4}
          maxLength={2000}
          placeholder="経験やアピールポイントがあれば記入してください"
        />
      </div>

      <Button type="submit" disabled={submitting}>
        {submitting ? '応募中…' : 'この案件に応募する'}
      </Button>

      {serverError && (
        <p role="alert" className="font-medium text-destructive">
          {serverError}
        </p>
      )}
    </form>
  )
}
