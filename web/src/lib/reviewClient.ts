// クライアント側から BFF（/api/contracts/{id}/reviews）を呼ぶ

export type ReviewResult = { ok: true } | { ok: false; error: string }

// レビューの投稿。提出＝確定で、編集も取り消しもできない
// （公開後に書き換えられると同時公開の意味が失われるため、api 側も更新を用意していない）
export async function submitReview(
  contractId: number,
  input: { rating: number; comment: string },
): Promise<ReviewResult> {
  let res: Response
  try {
    res = await fetch(`/api/contracts/${contractId}/reviews`, {
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

  if (!res.ok) {
    const json = await res.json().catch(() => null)
    return { ok: false, error: json?.error ?? 'エラーが発生しました' }
  }
  return { ok: true }
}
