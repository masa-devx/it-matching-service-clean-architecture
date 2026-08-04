// クライアント側から BFF（/api/profile）を呼ぶ。Go の URL もトークンも知らない

export type SaveResult = { ok: true } | { ok: false; error: string }

export async function saveProfile(input: unknown): Promise<SaveResult> {
  let res: Response
  try {
    res = await fetch('/api/profile', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input),
    })
  } catch {
    return {
      ok: false,
      error: '通信に失敗しました。時間をおいて再度お試しください',
    }
  }

  if (!res.ok) {
    const json = await res.json().catch(() => null)
    return { ok: false, error: json?.error ?? 'エラーが発生しました' }
  }
  return { ok: true }
}
