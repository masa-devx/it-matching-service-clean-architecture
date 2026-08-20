// Package e2efixture は API 統合テストの「基準世界」を提供する。
//
// 一次情報は seed.go（Go コード）で、dump.sql はそこから生成される生成物（手編集禁止）。
// テストは Load で基準世界をトランザクション内に読み込み、この ids.go の定数で参照する。
package e2efixture

// Password は基準世界の全ユーザー共通のログインパスワード。
// テストは本物の signin API にこの値を送り、JWT Cookie を得る（認証をモックしない）
const Password = "e2ePassword123"

// 以下の ID は「クリーンな DB に seed.go の順序で INSERT したときの連番」。
// seed.go は生成された ID がこの定数と一致することを検証するため、
// 世界の構築順を変えたのに定数を直し忘れると make e2e-dump の時点で失敗する（ズレたまま進めない）

// CompanyA は正常系の主役となる企業
var CompanyA = struct {
	UserID    int64
	CompanyID int64
	Email     string
}{UserID: 1, CompanyID: 1, Email: "company-a@example.com"}

// CompanyB は「他社」。所有チェック（404）・越境（403）の検証に使う
var CompanyB = struct {
	UserID    int64
	CompanyID int64
	Email     string
}{UserID: 2, CompanyID: 2, Email: "company-b@example.com"}

// TalentA / TalentB は人材2名。応募6状態を2名×公開案件3件で張るために2名必要
var TalentA = struct {
	UserID   int64
	TalentID int64
	Email    string
}{UserID: 3, TalentID: 1, Email: "talent-a@example.com"}

var TalentB = struct {
	UserID   int64
	TalentID int64
	Email    string
}{UserID: 4, TalentID: 2, Email: "talent-b@example.com"}

// Projects は案件5件。CompanyA が4件（draft / published×2 / closed）、CompanyB が公開1件
var Projects = struct {
	ADraft      int64 // CompanyA・下書き（talent からは見えない）
	APublished  int64 // CompanyA・公開中（応募 Applied / Offered が付く）
	APublished2 int64 // CompanyA・公開中（応募 Accepted / Rejected が付く）
	AClosed     int64 // CompanyA・募集終了
	BPublished  int64 // CompanyB・公開中（応募 Withdrawn / Declined が付く）
}{ADraft: 1, APublished: 2, APublished2: 3, AClosed: 4, BPublished: 5}

// Applications は応募の全6状態を1件ずつ。
// (talent, project) は UNIQUE のため、2名×公開3案件=6枠にちょうど1状態ずつ割り当てる
var Applications = struct {
	Applied   int64 // TalentA → APublished（応募したまま）
	Offered   int64 // TalentB → APublished（企業がオファー済み）
	Accepted  int64 // TalentA → APublished2（オファー→承諾＝ダブルオプトイン成立）
	Rejected  int64 // TalentB → APublished2（企業が不採用）
	Withdrawn int64 // TalentA → BPublished（人材が取り下げ）
	Declined  int64 // TalentB → BPublished（オファー→辞退）
}{Applied: 1, Offered: 2, Accepted: 3, Rejected: 4, Withdrawn: 5, Declined: 6}
