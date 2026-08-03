// クライアント側から Next.js の BFF（/api/auth/*）を呼ぶ関数群。
// Go API を直接呼ばない・トークンには一切触れない（httpOnly Cookie はブラウザが自動送信）

export type AuthResult = { ok: true } | { ok: false; error: string }

async function postJson(path: string, body: unknown): Promise<AuthResult> {
  let res: Response
  try {
    res = await fetch(path, {
      method: 'POST',
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

export function signup(input: {
  email: string
  password: string
  role: 'company' | 'talent'
}): Promise<AuthResult> {
  return postJson('/api/auth/signup', input)
}

export function login(input: {
  email: string
  password: string
}): Promise<AuthResult> {
  return postJson('/api/auth/login', input)
}

export function logout(): Promise<AuthResult> {
  return postJson('/api/auth/logout', {})
}
