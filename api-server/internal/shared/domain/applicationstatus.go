package domain

// ApplicationStatus は応募の選考状態（DB の applications.status と同じ語彙）
type ApplicationStatus string

const (
	ApplicationApplied   ApplicationStatus = "applied"   // 応募済み（初期状態）
	ApplicationOffered   ApplicationStatus = "offered"   // オファー中
	ApplicationAccepted  ApplicationStatus = "accepted"  // 承諾（ダブルオプトイン成立）
	ApplicationRejected  ApplicationStatus = "rejected"  // 不採用
	ApplicationWithdrawn ApplicationStatus = "withdrawn" // 取り下げ
	ApplicationDeclined  ApplicationStatus = "declined"  // 辞退
)

// AllApplicationStatuses は全状態の一覧（テストの検算・遷移先の列挙に使う）
var AllApplicationStatuses = []ApplicationStatus{
	ApplicationApplied,
	ApplicationOffered,
	ApplicationAccepted,
	ApplicationRejected,
	ApplicationWithdrawn,
	ApplicationDeclined,
}

// Actor は遷移を実行する側。auth のロールとは別に domain 内で定義する
// （domain は何にも依存しない。値が同じでも「認証上の役割」と「遷移の実行者」は別の概念）
type Actor string

const (
	ActorCompany Actor = "company"
	ActorTalent  Actor = "talent"
)

// AllActors は全実行者の一覧（テストの検算に使う）
var AllActors = []Actor{ActorCompany, ActorTalent}

type applicationTransition struct {
	actor Actor
	from  ApplicationStatus
	to    ApplicationStatus
}

// 許可される遷移だけを (誰が, どこから, どこへ) の3つ組で列挙する。
// 案件（projectstatus.go）との違いは2つ:
//   - actor も表の一部（オファーできるのは company だけ・承諾できるのは talent だけ）
//   - 終端がある（accepted / rejected / withdrawn / declined からはどこへも行けない）
//
// accepted へ到達する唯一の経路が applied →(company)→ offered →(talent)→ accepted
// ＝どちらか一方の意思だけでは成立しない（ダブルオプトイン）
var allowedApplicationTransitions = map[applicationTransition]bool{
	{actor: ActorCompany, from: ApplicationApplied, to: ApplicationOffered}:  true, // オファー
	{actor: ActorCompany, from: ApplicationApplied, to: ApplicationRejected}: true, // 不採用
	{actor: ActorTalent, from: ApplicationApplied, to: ApplicationWithdrawn}: true, // 取り下げ
	{actor: ActorTalent, from: ApplicationOffered, to: ApplicationAccepted}:  true, // 承諾
	{actor: ActorTalent, from: ApplicationOffered, to: ApplicationDeclined}:  true, // 辞退
	{actor: ActorTalent, from: ApplicationOffered, to: ApplicationWithdrawn}: true, // オファー後の取り下げ
}

// CanTransitApplication は actor による from → to が許可された遷移かを返す
func CanTransitApplication(actor Actor, from, to ApplicationStatus) bool {
	return allowedApplicationTransitions[applicationTransition{actor: actor, from: from, to: to}]
}

// ApplicationTransitionsFor は actor が from から遷移できる先を返す
// （エラーメッセージへの「可能な遷移」の埋め込み・画面のボタン出し分けに使う）
func ApplicationTransitionsFor(actor Actor, from ApplicationStatus) []ApplicationStatus {
	var tos []ApplicationStatus
	for _, to := range AllApplicationStatuses {
		if CanTransitApplication(actor, from, to) {
			tos = append(tos, to)
		}
	}
	return tos
}
