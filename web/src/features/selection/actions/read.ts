'use server'

import { listApplications } from '@/external/handler/selection'

// queryFn の通り道（#44 以降の型）。所有チェックは API の JOIN 越し WHERE が行う
export async function fetchApplications(projectId: number) {
  return listApplications(projectId)
}
