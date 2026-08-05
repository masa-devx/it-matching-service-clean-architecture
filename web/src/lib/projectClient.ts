// クライアントから BFF（/api/projects）を呼ぶ。Go の URL もトークンも知らない

export type CreateProjectResult =
  { ok: true; id: number } | { ok: false; error: string }

export async function createProject(
  input: unknown,
): Promise<CreateProjectResult> {
  let res: Response
  try {
    res = await fetch('/api/projects', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
    })
  } catch {
    return {
      ok: false,
      error: '通信に失敗しました。時間をおいて再度お試しください',
    }
  }

  const json = await res.json().catch(() => null)
  if (!res.ok) {
    return { ok: false, error: json?.error ?? 'エラーが発生しました' }
  }
  return { ok: true, id: json?.id }
}
