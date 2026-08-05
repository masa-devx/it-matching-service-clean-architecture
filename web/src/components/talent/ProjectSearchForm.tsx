'use client'

import { useRouter, useSearchParams } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import type { ProjectSearchParams } from '@/lib/projects'

// 検索条件は state ではなく URL に持たせる。
// リロード・共有・ブラウザバックで同じ結果に戻れ、RSC が条件を直接受け取れる
export function ProjectSearchForm({
  defaultValues,
}: {
  defaultValues: ProjectSearchParams
}) {
  const router = useRouter()
  const searchParams = useSearchParams()

  function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    const params = new URLSearchParams()

    for (const key of [
      'skills',
      'rate_min',
      'rate_max',
      'hours_max',
      'q',
    ] as const) {
      const value = String(formData.get(key) ?? '').trim()
      // 空欄は URL に残さない（条件が増えるほどURLが読みづらくなるため）
      if (value) {
        params.set(key, value)
      }
    }
    if (formData.get('remote_ok') === 'on') {
      params.set('remote_ok', 'true')
    }
    // 条件を変えたらページは1に戻す（3ページ目のまま絞り込むと空になりうる）
    router.push(params.toString() ? `?${params.toString()}` : '?')
  }

  const hasConditions = Array.from(searchParams.keys()).some(
    (key) => key !== 'page',
  )

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-4 rounded-lg border bg-card p-4"
    >
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="flex flex-col gap-2 sm:col-span-2">
          <Label htmlFor="skills">スキル</Label>
          <Input
            id="skills"
            name="skills"
            className="h-11"
            placeholder="Go, React（カンマ区切り・すべて含む）"
            defaultValue={defaultValues.skills ?? ''}
          />
        </div>

        <div className="flex flex-col gap-2 sm:col-span-2">
          <Label htmlFor="q">キーワード</Label>
          <Input
            id="q"
            name="q"
            className="h-11"
            placeholder="案件タイトルに含まれる語"
            defaultValue={defaultValues.q ?? ''}
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="rate_min">時給の下限（円）</Label>
          <Input
            id="rate_min"
            name="rate_min"
            type="number"
            inputMode="numeric"
            className="h-11"
            placeholder="4000"
            defaultValue={defaultValues.rate_min ?? ''}
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="rate_max">時給の上限（円）</Label>
          <Input
            id="rate_max"
            name="rate_max"
            type="number"
            inputMode="numeric"
            className="h-11"
            placeholder="10000"
            defaultValue={defaultValues.rate_max ?? ''}
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="hours_max">週の稼働上限（時間）</Label>
          <Input
            id="hours_max"
            name="hours_max"
            type="number"
            inputMode="numeric"
            className="h-11"
            placeholder="20"
            defaultValue={defaultValues.hours_max ?? ''}
          />
        </div>

        <label className="flex items-center gap-2 self-end pb-3 text-sm">
          <input
            type="checkbox"
            name="remote_ok"
            className="size-4 accent-primary"
            defaultChecked={defaultValues.remote_ok === 'true'}
          />
          フルリモート可のみ
        </label>
      </div>

      <div className="flex gap-3">
        <Button type="submit" className="h-11">
          この条件で検索
        </Button>
        {hasConditions && (
          <Button
            type="button"
            variant="outline"
            className="h-11"
            onClick={() => router.push('?')}
          >
            条件をクリア
          </Button>
        )}
      </div>
    </form>
  )
}
