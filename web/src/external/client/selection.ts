import 'server-only'

import {
  applicationsOffer,
  applicationsReject,
  projectsListApplications,
} from '@repo/api-client/company/generated/endpoints'

import { authHeaders } from './auth'

// 選考（company 視点: 応募を受け取る・選考する）のクライアント
export const selectionClient = {
  listApplications: async (projectId: number) =>
    projectsListApplications(projectId, { headers: await authHeaders() }),
  // メッセージは任意（画面からの入力は後続対応。契約変更で生成クライアントに body 引数が増えた）
  offer: async (applicationId: number, message?: string) =>
    applicationsOffer(applicationId, message ? { message } : undefined, {
      headers: await authHeaders(),
    }),
  reject: async (applicationId: number) =>
    applicationsReject(applicationId, { headers: await authHeaders() }),
}
