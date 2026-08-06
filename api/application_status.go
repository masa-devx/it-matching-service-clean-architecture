package main

import "slices"

// 選考の状態。DBのCHECK制約（applications.status）と対応させる
const (
	applicationStatusApplied   = "applied"
	applicationStatusOffered   = "offered"
	applicationStatusAccepted  = "accepted"
	applicationStatusRejected  = "rejected"
	applicationStatusWithdrawn = "withdrawn"
	applicationStatusDeclined  = "declined"
)

// applicationStatuses は取りうる状態の一覧。DBのCHECK制約と対応させる。
// 絞り込みクエリの検証に使う（存在しない状態を指定されたら400にするため）
var applicationStatuses = []string{
	applicationStatusApplied,
	applicationStatusOffered,
	applicationStatusAccepted,
	applicationStatusRejected,
	applicationStatusWithdrawn,
	applicationStatusDeclined,
}

// isApplicationStatus は文字列が状態として定義済みかを返す
func isApplicationStatus(s string) bool {
	return slices.Contains(applicationStatuses, s)
}

// statusTransition は「どの状態から、どの状態へ」の1手。
type statusTransition struct {
	from string
	to   string
}

// applicationTransitions は許可される状態遷移とその実行者。状態機械①の唯一の定義。
//
// ここに無い組み合わせはすべて禁止（終端状態は from に現れないので自動的に遷移不可）。
// 遷移の判定をハンドラに散らすと「ある経路からだけ不正な状態になれる」バグが生まれるため、
// 追加・変更は必ずこの表だけを書き換える。
//
// accepted への行が「offered から・人材のみ」の1つしかないことが、
// ダブルオプトイン（企業のオファーと人材の承諾が揃って初めて成立する）の実装そのもの。
var applicationTransitions = map[statusTransition]string{
	{from: applicationStatusApplied, to: applicationStatusOffered}:   roleCompany, // オファー
	{from: applicationStatusApplied, to: applicationStatusRejected}:  roleCompany, // 見送り
	{from: applicationStatusApplied, to: applicationStatusWithdrawn}: roleTalent,  // 取り下げ
	{from: applicationStatusOffered, to: applicationStatusAccepted}:  roleTalent,  // 承諾＝成立
	{from: applicationStatusOffered, to: applicationStatusDeclined}:  roleTalent,  // 辞退
	{from: applicationStatusOffered, to: applicationStatusWithdrawn}: roleTalent,  // 取り下げ
}

// canTransition は from から to への遷移を role が実行してよいかを返す。
// 状態と実行者の両方が一致しなければ許可しない（企業が単独で accepted にできない）
func canTransition(from, to, role string) bool {
	actor, ok := applicationTransitions[statusTransition{from: from, to: to}]
	return ok && actor == role
}
