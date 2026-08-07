// クライアント側から BFF（/api/contracts/{id}/messages）を呼ぶ

export type MessageResult = { ok: true } | { ok: false; error: string }

export async function sendMessage(
  contractId: number,
  body: string,
): Promise<MessageResult> {
  let res: Response
  try {
    res = await fetch(`/api/contracts/${contractId}/messages`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ body }),
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
