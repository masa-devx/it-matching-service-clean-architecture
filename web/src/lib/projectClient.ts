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

export type ProjectResult = { ok: true } | { ok: false; error: string }

export async function updateProject(
  id: number,
  input: unknown,
): Promise<ProjectResult> {
  return send(`/api/projects/${id}`, 'PUT', input)
}

// 遷移が許されるかは Go の遷移表が判定する。
// 画面はボタンを出し分けるだけで、正しさの担保はしない（409 が最後の砦）
export async function updateProjectStatus(
  id: number,
  status: string,
): Promise<ProjectResult> {
  return send(`/api/projects/${id}/status`, 'PATCH', { status })
}

async function send(
  url: string,
  method: 'PUT' | 'PATCH',
  body: unknown,
): Promise<ProjectResult> {
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
