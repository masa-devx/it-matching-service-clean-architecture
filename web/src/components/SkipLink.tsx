// キーボード利用者が毎回ヘッダーのナビを通過せずに本文へ移動できるようにする。
// 普段は視覚的に隠し（sr-only）、Tab でフォーカスが当たったときだけ表示する
export function SkipLink() {
  return (
    <a
      href="#main"
      className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:rounded-md focus:bg-primary focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-primary-foreground"
    >
      本文へスキップ
    </a>
  )
}
