package main

import "slices"

// 稼働報告の状態。DBのCHECK制約（work_reports.status）と対応させる
const (
	workReportStatusSubmitted = "submitted" // 提出済み（企業の確認待ち）
	workReportStatusApproved  = "approved"  // 承認
	workReportStatusRejected  = "rejected"  // 差し戻し
)

// workReportStatuses は取りうる状態の一覧。テストの網羅を検算するために使う
var workReportStatuses = []string{
	workReportStatusSubmitted,
	workReportStatusApproved,
	workReportStatusRejected,
}

// workReportTransitions は許可される稼働報告の状態遷移と、それを実行できるロール。
//
// 契約の遷移表（contractTransitions）と同じく値は []string だが、
// こちらは実際には全行が1ロールずつ。型を揃えているのは、状態遷移の判定を
// 同じ形で書けるようにするため（読む側が別の作法を覚えなくて済む）。
//
// rejected → submitted（再提出）が「戻る遷移」。契約の差し戻しと対になっており、
// 「企業が差し戻す → 人材が直して出し直す」という往復を何度でも繰り返せる。
//
// approved は終端。いったん承認した報告を後から覆せるようにすると、
// 「承認済み＝確定した実績」という前提が崩れ、報酬計算の根拠にできなくなる
var workReportTransitions = map[statusTransition][]string{
	// 確認するのは企業。人材が自分の報告を承認できてしまうと、稼働報告が
	// 「第三者に確認された記録」ではなく単なる自己申告になってしまう
	{from: workReportStatusSubmitted, to: workReportStatusApproved}: {roleCompany},
	{from: workReportStatusSubmitted, to: workReportStatusRejected}: {roleCompany},

	// 差し戻された報告を直して出し直すのは人材
	{from: workReportStatusRejected, to: workReportStatusSubmitted}: {roleTalent},
}

// canTransitionWorkReport は from から to への遷移を role が実行してよいかを返す
func canTransitionWorkReport(from, to, role string) bool {
	actors, ok := workReportTransitions[statusTransition{from: from, to: to}]
	return ok && slices.Contains(actors, role)
}
