// クライアント側から BFF を呼ぶ。Go の URL もトークンも知らない

export type ContractResult = { ok: true } | { ok: false; error: string }

// 遷移が許されるかは Go の遷移表が判定する。
// 画面はボタンを出し分けるだけで、正しさの担保はしない（409 が最後の砦）
export async function updateContractStatus(
  contractId: number,
  status: string,
): Promise<ContractResult> {
  return send(`/api/contracts/${contractId}/status`, 'PATCH', { status })
}

export type WorkReportInput = {
  week_start: string
  hours: number
  summary: string
}

// 稼働報告の新規提出。week_start は週内のどの日でもよく、サーバー側で月曜に丸められる
export async function submitWorkReport(
  contractId: number,
  input: WorkReportInput,
): Promise<ContractResult> {
  return send(`/api/contracts/${contractId}/work-reports`, 'POST', input)
}

// 差し戻された報告の修正＋再提出。
// 内容の更新と状態の復帰（rejected → submitted）を1操作で行うのは、
// 稼働報告に下書きの概念が無く「内容を出すこと自体が提出」だから（api 側も同じ設計）
export async function resubmitWorkReport(
  workReportId: number,
  input: Omit<WorkReportInput, 'week_start'>,
): Promise<ContractResult> {
  return send(`/api/work-reports/${workReportId}`, 'PUT', input)
}

// 稼働報告の承認・差し戻し（企業）。
// 差し戻しでは理由が必須だが、その検証は Go 側が行う。
// 画面では「理由が空なら実行ボタンを押せない」ようにして、送信前に気づけるようにする
export async function reviewWorkReport(
  workReportId: number,
  status: 'approved' | 'rejected',
  reviewNote: string,
): Promise<ContractResult> {
  return send(`/api/work-reports/${workReportId}/status`, 'PATCH', {
    status,
    review_note: reviewNote,
  })
}

async function send(
  url: string,
  method: 'POST' | 'PUT' | 'PATCH',
  body: unknown,
): Promise<ContractResult> {
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
