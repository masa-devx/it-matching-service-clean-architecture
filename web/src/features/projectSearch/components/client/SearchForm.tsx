'use client'

import { Search } from 'lucide-react'
import { useRouter } from 'next/navigation'

import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

import type { ProjectSearchFilters } from '../../queries/projects'

// 検索状態の一次情報は URL。このフォームは「URL を編集する道具」でしかなく、
// 自前の state を持たない（uncontrolled）。submit で URL を書き換えると
// RSC が再実行され、新しい条件で prefetch → 一覧が切り替わる
export function SearchForm({
  defaultValues,
}: {
  defaultValues: ProjectSearchFilters
}) {
  const router = useRouter()

  const onSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const form = new FormData(e.currentTarget)
    const params = new URLSearchParams()

    const skills = String(form.get('skills') ?? '')
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    if (skills.length > 0) {
      params.set('skills', skills.join(','))
    }
    // Radix の Checkbox は form 内では hidden input を描画し、チェック時のみ値が載る
    if (form.get('remote_ok')) {
      params.set('remote_ok', 'true')
    }
    const rate = String(form.get('min_hourly_rate') ?? '').trim()
    if (rate !== '') {
      params.set('min_hourly_rate', rate)
    }

    const query = params.toString()
    router.push(query ? `/talent/projects?${query}` : '/talent/projects')
  }

  return (
    <Card>
      <CardContent>
        <form onSubmit={onSubmit} className="flex flex-wrap items-end gap-4">
          <div className="space-y-1.5">
            <Label htmlFor="skills">スキル（カンマ区切り・AND）</Label>
            <Input
              id="skills"
              name="skills"
              placeholder="Go, PostgreSQL"
              defaultValue={defaultValues.skills.join(', ')}
              className="w-56"
            />
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="min_hourly_rate">時給下限（円）</Label>
            <Input
              id="min_hourly_rate"
              name="min_hourly_rate"
              type="number"
              min={1}
              defaultValue={defaultValues.minHourlyRate ?? ''}
              className="w-32"
            />
          </div>

          <div className="flex h-8 items-center gap-2">
            <Checkbox
              id="remote_ok"
              name="remote_ok"
              defaultChecked={defaultValues.remoteOk === true}
            />
            <Label htmlFor="remote_ok">リモート可のみ</Label>
          </div>

          <div className="flex gap-2">
            <Button type="submit">
              <Search data-icon="inline-start" aria-hidden="true" />
              検索
            </Button>
            <Button
              type="button"
              variant="ghost"
              onClick={() => router.push('/talent/projects')}
            >
              クリア
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
