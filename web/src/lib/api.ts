import 'server-only'

// Go API への呼び出しはこのファイルに集約する（コンポーネントに fetch を書かない）。
// "server-only" により、クライアントコンポーネントから import するとビルドエラーになる
// （トークンを扱う処理がブラウザに漏れる事故をコンパイル時に防ぐ）

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8082'

export type ApiError = {
  status: number
  message: string
}

type ApiResult<T> =
  { data: T; error?: never } | { data?: never; error: ApiError }

export async function apiPost<T>(
  path: string,
  body: unknown,
  token?: string,
): Promise<ApiResult<T>> {
  return request<T>(path, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(body),
  })
}

// token を渡すと Authorization: Bearer で Go の保護ルートを呼べる
export async function apiGet<T>(
  path: string,
  token?: string,
): Promise<ApiResult<T>> {
  return request<T>(path, {
    method: 'GET',
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
}

export async function apiPut<T>(
  path: string,
  body: unknown,
  token?: string,
): Promise<ApiResult<T>> {
  return request<T>(path, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(body),
  })
}

// PATCH は「リソースの一部を更新する」意味。応募の状態更新のように
// 全体を送り直さない更新に使う（PUT は全体置換）
export async function apiPatch<T>(
  path: string,
  body: unknown,
  token?: string,
): Promise<ApiResult<T>> {
  return request<T>(path, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify(body),
  })
}

async function request<T>(
  path: string,
  init: RequestInit,
): Promise<ApiResult<T>> {
  let res: Response
  try {
    res = await fetch(`${API_URL}${path}`, { ...init, cache: 'no-store' })
  } catch {
    // ネットワークレベルの失敗（APIサーバー未起動など）
    return {
      error: {
        status: 0,
        message: 'サーバーに接続できません。時間をおいて再度お試しください',
      },
    }
  }

  const json = await res.json().catch(() => null)
  if (!res.ok) {
    return {
      error: {
        status: res.status,
        // Go API は { "error": "メッセージ" } を返す（response.go の writeError）
        message: json?.error ?? 'エラーが発生しました',
      },
    }
  }
  return { data: json as T }
}
