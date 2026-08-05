// 必須項目の印。記号だけでは意味が伝わらないため、スクリーンリーダー向けに
// 「必須」と読み上げるテキストを添える
export function RequiredMark() {
  return (
    <span className="text-destructive">
      <span aria-hidden="true">*</span>
      <span className="sr-only">必須</span>
    </span>
  )
}
