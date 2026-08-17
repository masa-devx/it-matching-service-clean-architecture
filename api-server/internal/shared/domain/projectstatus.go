// Package domain は業務ルール（状態遷移・不変条件）を置く。
// DB にも HTTP にも依存しない純粋なパッケージ（depguard の domain-purity で強制）。
// 遷移を持つ状態は、ここに「許可の表（ホワイトリスト）」として一元定義する
package domain

// ProjectStatus は案件の掲載状態（DB の projects.status と同じ語彙）
type ProjectStatus string

const (
	ProjectDraft     ProjectStatus = "draft"     // 下書き
	ProjectPublished ProjectStatus = "published" // 公開中
	ProjectClosed    ProjectStatus = "closed"    // 募集終了
)

// AllProjectStatuses は全状態の一覧（テストの検算・遷移先の列挙に使う）
var AllProjectStatuses = []ProjectStatus{ProjectDraft, ProjectPublished, ProjectClosed}

type projectTransition struct {
	from ProjectStatus
	to   ProjectStatus
}

// 許可される遷移だけを列挙する。表に無い組み合わせは自動的に拒否される（ホワイトリスト）。
// 「終端が無い」のがこの状態機械の特徴（closed からも再募集で戻れる）
var allowedProjectTransitions = map[projectTransition]bool{
	{from: ProjectDraft, to: ProjectPublished}:  true, // 公開
	{from: ProjectPublished, to: ProjectDraft}:  true, // 非公開（下書きへ戻す）
	{from: ProjectPublished, to: ProjectClosed}: true, // 募集終了
	{from: ProjectClosed, to: ProjectPublished}: true, // 再募集
}

// CanTransitProject は from → to が許可された遷移かを返す
func CanTransitProject(from, to ProjectStatus) bool {
	return allowedProjectTransitions[projectTransition{from: from, to: to}]
}

// ProjectTransitionsFrom は from から遷移できる先を返す
// （エラーメッセージへの「可能な遷移」の埋め込み・画面のボタン出し分けに使う）
func ProjectTransitionsFrom(from ProjectStatus) []ProjectStatus {
	var tos []ProjectStatus
	for _, to := range AllProjectStatuses {
		if CanTransitProject(from, to) {
			tos = append(tos, to)
		}
	}
	return tos
}
