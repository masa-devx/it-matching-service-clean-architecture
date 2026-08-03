import 'server-only'

// Go API への呼び出しはこのファイルに集約する（コンポーネントに fetch を書かない）。
// "server-only" により、クライアントコンポーネントから import するとビルドエラーになる
// （トークンを扱う処理がブラウザに漏れる事故をコンパイル時に防ぐ）

const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8081'

export type ApiError = {
  status: number
  message: string
}

type ApiResult<T> =
  | { data: T; error?: never }
  | { data?: never; error: ApiError }

export async function apiPost<T>(
  path: string,
  body: unknown,
): Promise<ApiResult<T>> {
  let res: Response
  try {
    res = await fetch(`${API_URL}${path}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      cache: 'no-store',
    })
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
