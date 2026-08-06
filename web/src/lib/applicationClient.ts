// クライアント側から BFF（/api/applications）を呼ぶ。Go の URL もトークンも知らない

export type ApplicationResult = { ok: true } | { ok: false; error: string }

export async function applyToProject(
  projectId: number,
  message: string,
): Promise<ApplicationResult> {
  return send('/api/applications', 'POST', {
    project_id: projectId,
    message,
  })
}

// 状態の遷移が許されるかは Go の遷移表が判定する。
// 画面はボタンを出し分けるだけで、正しさの担保はしない（409 が最後の砦）
export async function updateApplicationStatus(
  applicationId: number,
  status: string,
): Promise<ApplicationResult> {
  return send(`/api/applications/${applicationId}/status`, 'PATCH', {
    status,
  })
}

async function send(
  url: string,
  method: 'POST' | 'PATCH',
  body: unknown,
): Promise<ApplicationResult> {
  let res: Response
  try {
    res = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
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
