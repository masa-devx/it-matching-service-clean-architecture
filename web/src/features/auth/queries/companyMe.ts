import { queryOptions } from '@tanstack/react-query'

import { fetchCompanyMe } from '../actions/me'

// queries/ はクエリキーと取得方法（queryOptions）の一元管理場所。
// キーをリテラルで散らばらせると invalidate の指定ミスが起きるため、
// 「このデータのキーはここにしか書かれていない」状態を保つ
export const companyMeQuery = queryOptions({
  queryKey: ['auth', 'company', 'me'] as const,
  queryFn: () => fetchCompanyMe(),
})
