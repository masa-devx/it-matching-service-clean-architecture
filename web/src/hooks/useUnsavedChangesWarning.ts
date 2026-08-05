'use client'

import { useEffect } from 'react'

// 入力途中でタブを閉じる・リロードするときに確認ダイアログを出す。
//
// 制約: 表示される文言はブラウザ固定で、カスタムメッセージは指定できない
// （スパム的な引き止めを防ぐための仕様）。
// また Next.js のクライアント遷移（Link）は beforeunload では捕捉できないため、
// ここではブラウザ操作による離脱のみを対象にしている
export function useUnsavedChangesWarning(isDirty: boolean) {
  useEffect(() => {
    if (!isDirty) {
      return
    }

    function handleBeforeUnload(e: BeforeUnloadEvent) {
      e.preventDefault()
      // 一部ブラウザは returnValue の設定を要求する（値自体は表示されない）
      e.returnValue = ''
    }

    window.addEventListener('beforeunload', handleBeforeUnload)
    return () => window.removeEventListener('beforeunload', handleBeforeUnload)
  }, [isDirty])
}
