import 'server-only'

import { cache } from 'react'

import type {
  TsunaguWorksCompanyMe,
  TsunaguWorksCompanySignupInput,
  TsunaguWorksLoginInput,
} from '@repo/api-client/company/generated/models'
import type {
  TsunaguWorksTalentMe,
  TsunaguWorksTalentSignupInput,
} from '@repo/api-client/talent/generated/models'

import { companyClient, talentClient } from '../client/auth'
import {
  clearSessionToken,
  getSessionToken,
  setSessionToken,
} from '../client/session'

// external/handler は「入口」: features の actions から呼ばれ、
// 通信（client）と Cookie の操作をまとめてエラーを画面向けの形に整える

type Result = { ok: true } | { ok: false; error: string }

const CONNECT_ERROR = 'サーバーに接続できませんでした'

export async function signupCompany(
  input: TsunaguWorksCompanySignupInput,
): Promise<Result> {
  try {
    const res = await companyClient.signup(input)
    if (res.status !== 201) {
      return {
        ok: false,
        error: res.data.error ?? 'サインアップに失敗しました',
      }
    }
    await setSessionToken(res.data.token)
    return { ok: true }
  } catch {
    return { ok: false, error: CONNECT_ERROR }
  }
}

export async function loginCompany(
  input: TsunaguWorksLoginInput,
): Promise<Result> {
  try {
    const res = await companyClient.login(input)
    if (res.status !== 200) {
      return { ok: false, error: res.data.error ?? 'ログインに失敗しました' }
    }
    await setSessionToken(res.data.token)
    return { ok: true }
  } catch {
    return { ok: false, error: CONNECT_ERROR }
  }
}

export async function signupTalent(
  input: TsunaguWorksTalentSignupInput,
): Promise<Result> {
  try {
    const res = await talentClient.signup(input)
    if (res.status !== 201) {
      return {
        ok: false,
        error: res.data.error ?? 'サインアップに失敗しました',
      }
    }
    await setSessionToken(res.data.token)
    return { ok: true }
  } catch {
    return { ok: false, error: CONNECT_ERROR }
  }
}

export async function loginTalent(
  input: TsunaguWorksLoginInput,
): Promise<Result> {
  try {
    const res = await talentClient.login(input)
    if (res.status !== 200) {
      return { ok: false, error: res.data.error ?? 'ログインに失敗しました' }
    }
    await setSessionToken(res.data.token)
    return { ok: true }
  } catch {
    return { ok: false, error: CONNECT_ERROR }
  }
}

export async function logout(): Promise<void> {
  await clearSessionToken()
}

// me はガード（layout）と画面表示が使う。未認証・通信不能は null に潰す
// （呼び出し側は「ログイン済みか否か」だけを知ればよい）
export async function meCompany(): Promise<TsunaguWorksCompanyMe | null> {
  try {
    const res = await companyClient.me()
    return res.status === 200 ? res.data : null
  } catch {
    return null
  }
}

export async function meTalent(): Promise<TsunaguWorksTalentMe | null> {
  try {
    const res = await talentClient.me()
    return res.status === 200 ? res.data : null
  } catch {
    return null
  }
}

// currentRole は (guest) ガードと LP の導線出し分けが使う: ログイン済みならどちらのロールかを返す。
// トークンが無ければ API を呼ばずに即 null（ゲストページの表示を遅くしない）。
// React の cache() で「同一リクエスト内なら layout と page で呼んでも API は1回」にする
export const currentRole = cache(
  async (): Promise<'company' | 'talent' | null> => {
    const token = await getSessionToken()
    if (!token) {
      return null
    }
    if (await meCompany()) {
      return 'company'
    }
    if (await meTalent()) {
      return 'talent'
    }
    return null
  },
)
