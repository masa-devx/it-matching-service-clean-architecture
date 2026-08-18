'use server'

import { listMyApplications } from '@/external/handler/application'

// queryFn の通り道 兼 RSC の応募済み判定用（案件詳細ページが自分の応募を照合する）
export async function fetchMyApplications() {
  return listMyApplications()
}
