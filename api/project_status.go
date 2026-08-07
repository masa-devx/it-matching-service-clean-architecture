package main

// projectTransitions は許可される掲載状態の遷移。状態機械③の唯一の定義。
//
// 応募の遷移表（applicationTransitions）と違い、値がロールではなく bool なのは、
// 案件を操作できるのが常に所有企業だけだから（誰が実行できるかを表に持つ必要がない）。
// 終端状態も無い：closed からも published に戻せる（採用が流れた場合の再募集や、
// M2 で契約が中止されたときに募集を再開するため）。
//
// draft → closed / closed → draft は許可しない。
// 「一度も公開していない案件を終了する」「終了した案件を下書きに戻す」は意味が不明瞭なため
var projectTransitions = map[statusTransition]bool{
	{from: projectStatusDraft, to: projectStatusPublished}:  true, // 公開する
	{from: projectStatusPublished, to: projectStatusDraft}:  true, // 非公開に戻す
	{from: projectStatusPublished, to: projectStatusClosed}: true, // 募集を終了する
	{from: projectStatusClosed, to: projectStatusPublished}: true, // 再募集する
}

// canTransitionProject は from から to への掲載状態の変更が許可されるかを返す。
// 表に無い組み合わせ（同じ状態への変更を含む）はすべて false
func canTransitionProject(from, to string) bool {
	return projectTransitions[statusTransition{from: from, to: to}]
}
