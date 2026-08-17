import { describe, expect, it } from 'vitest'

import { projectFormSchema } from './create'

// 目的: 「生成Zodを .omit().extend() で加工する」方針（ADR-0007 案1）が成立することを実行で保証する。
//       このテストが通る＝案1採用の根拠（壊れたら方針の見直しシグナル）
// 観点: omit+extend の合成 / transform（カンマ区切り→配列）/ 生成側の制約（maxLength）の引き継ぎ /
//       API に存在しないフィールドを送る形が無いこと
describe('projectFormSchema（生成Zodの加工）', () => {
  const validInput = {
    title: 'Go APIの改修',
    description: '詳細です',
    hours_per_week: 10,
    remote_ok: true,
    required_skills: 'Go, PostgreSQL, ,TypeScript',
  }

  it('カンマ区切りの required_skills が配列に変換される（空要素は除去）', () => {
    const result = projectFormSchema.parse(validInput)
    expect(result.required_skills).toEqual(['Go', 'PostgreSQL', 'TypeScript'])
  })

  it('生成スキーマ由来の制約が引き継がれる（title 101文字は拒否）', () => {
    const result = projectFormSchema.safeParse({
      ...validInput,
      title: 'あ'.repeat(101),
    })
    expect(result.success).toBe(false)
  })

  it('optional のフィールド（hourly_rate_min）は未入力でも通る', () => {
    const result = projectFormSchema.safeParse(validInput)
    expect(result.success).toBe(true)
  })

  it('status を持たない（parse 結果に現れない＝公開状態を指定できる形が存在しない）', () => {
    const result = projectFormSchema.parse({
      ...validInput,
      status: 'published',
    })
    expect('status' in result).toBe(false)
  })
})
