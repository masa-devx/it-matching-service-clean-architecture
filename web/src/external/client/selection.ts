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
  offer: async (applicationId: number) =>
    applicationsOffer(applicationId, { headers: await authHeaders() }),
  reject: async (applicationId: number) =>
    applicationsReject(applicationId, { headers: await authHeaders() }),
}
