'use client'

import { logoutAction } from '../../actions/logout'

type Props = {
  role: 'company' | 'talent'
}

export function LogoutButton({ role }: Props) {
  return (
    <button
      type="button"
      onClick={() => logoutAction(role)}
      className="rounded border px-3 py-1 text-sm hover:bg-gray-100"
    >
      ログアウト
    </button>
  )
}
