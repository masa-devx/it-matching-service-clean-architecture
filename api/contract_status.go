package main

import "slices"

// 契約の状態。DBのCHECK制約（contracts.status）と対応させる
const (
	contractStatusActive    = "active"    // 成立（合意したが稼働前）
	contractStatusWorking   = "working"   // 稼働中
	contractStatusReviewing = "reviewing" // 検収待ち
	contractStatusCompleted = "completed" // 完了
	contractStatusCancelled = "cancelled" // 中止
)

// contractStatuses は取りうる状態の一覧。
// テストで「全組み合わせを網羅したか」を検算するために使う
// （絞り込みクエリの検証に使う isContractStatus は、必要になる #106 で追加する）
var contractStatuses = []string{
	contractStatusActive,
	contractStatusWorking,
	contractStatusReviewing,
	contractStatusCompleted,
	contractStatusCancelled,
}

// contractTransitions は許可される契約の状態遷移と、それを実行できるロール。状態機械②の唯一の定義。
//
// 値が []string（ロールの集合）なのは、中止だけが企業・人材の両方から実行できるため。
// 応募の遷移表（applicationTransitions）は1遷移＝1ロールだったので string で足りたが、
// ここでは「誰が実行できるか」が遷移によって1人だったり2人だったりする。
// 単一ロールの行が多いのに配列にしているのは、中止の3行のために型を揃える必要があるから。
//
// 設計の要点は2つ:
//
//  1. reviewing → working（差し戻し）が、このプロジェクトで初めての「戻る遷移」。
//     検収で不備が見つかったときに稼働へ戻す。何度でも往復しうる（検収→差し戻し→検収…）。
//     戻る遷移があるため、状態が単調に進むことを前提にした処理を書いてはいけない
//     （例: started_at を working に入るたび上書きすると、差し戻しのたび開始日が変わってしまう）。
//
//  2. 中止は active / working / reviewing のいずれからも、どちらの当事者からも実行できる。
//     取引は双方の合意で始まるが、続けられなくなる事情はどちらにも起こりうるため。
//     ただし completed（完了）からは中止できない——検収が済んだ取引を後から無かったことにはしない。
var contractTransitions = map[statusTransition][]string{
	// 稼働の開始は人材が宣言する。企業が「働き始めたことにする」のは実態と合わない
	{from: contractStatusActive, to: contractStatusWorking}: {roleTalent},

	// 検収の依頼は「作業が終わった」という人材側の申告
	{from: contractStatusWorking, to: contractStatusReviewing}: {roleTalent},

	// 検収の可否は企業が判断する。人材が自分で完了させられないことが、
	// 「成果を確認してから完了する」という検収の意味そのもの
	{from: contractStatusReviewing, to: contractStatusCompleted}: {roleCompany},
	{from: contractStatusReviewing, to: contractStatusWorking}:   {roleCompany}, // 差し戻し

	// 中止は双方から。完了前のどの段階でも起こりうる
	{from: contractStatusActive, to: contractStatusCancelled}:    {roleCompany, roleTalent},
	{from: contractStatusWorking, to: contractStatusCancelled}:   {roleCompany, roleTalent},
	{from: contractStatusReviewing, to: contractStatusCancelled}: {roleCompany, roleTalent},
}

// canTransitionContract は from から to への遷移を role が実行してよいかを返す。
// 表に無い組み合わせ（終端状態からの遷移を含む）はすべて false
func canTransitionContract(from, to, role string) bool {
	actors, ok := contractTransitions[statusTransition{from: from, to: to}]
	return ok && slices.Contains(actors, role)
}
