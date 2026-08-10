'use server'

import { projectsCreate } from '@repo/api-client/company/generated/endpoints'
import type { TsunaguWorksProject } from '@repo/api-client/company/generated/models'

import type { ProjectFormOutput } from './schema'

type ActionResult =
  { ok: true; project: TsunaguWorksProject } | { ok: false; error: string }

// Server Action: フォームの確定値（API の形）を受け取り、生成クライアントで Go API を呼ぶ。
// トークン等の秘匿情報はサーバー側だけで扱える（Phase 2 で認証ヘッダーがここに乗る）
export async function createProjectAction(
  input: ProjectFormOutput,
): Promise<ActionResult> {
  try {
    const res = await projectsCreate(input)
    if (res.status !== 201) {
      return { ok: false, error: res.data.error ?? '案件の作成に失敗しました' }
    }
    return { ok: true, project: res.data }
  } catch {
    // 接続失敗など。詳細はサーバーログに任せ、画面には安全な文言だけを返す
    return { ok: false, error: 'サーバーに接続できませんでした' }
  }
}
