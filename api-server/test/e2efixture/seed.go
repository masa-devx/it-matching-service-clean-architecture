package e2efixture

import (
	"context"
	"fmt"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	companyuc "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/domain"
	talentuc "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/talent/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/factories"
)

// passwordHash は Password の bcrypt ハッシュ。
// 毎回 bcrypt.GenerateFromPassword すると乱数ソルトで dump.sql が毎回変わり、
// 「同じ seed からは同じ dump」という再現性（CI 差分チェックの前提）が壊れるため固定値にする
// テスト専用の公開値（Password 定数のハッシュ）であり秘匿情報ではないため G101 を除外する
const passwordHash = "$2a$10$OW9eXMdehG7ljy2gSTWfleDRMTG57B0QRv6im2FtyPaBp5Ad1UifS" //nolint:gosec

// Seed は基準世界を構築する。クリーンな（マイグレーション適用済み・データ空の）DB を前提とする。
//
// 方針: 状態機械を持つデータ（projects.status / applications.status）は usecase＝遷移表を
// 通して作る（アプリが作り得ない状態を世界に混入させない）。状態を持たないマスタ
// （users / profiles）は factories で直接作る（bcrypt の都合もこちらで吸収する）
func Seed(ctx context.Context, dbtx db.DBTX) error {
	queries := db.New(dbtx)

	if err := seedUsers(ctx, queries); err != nil {
		return err
	}
	if err := seedProjects(ctx, queries); err != nil {
		return err
	}
	if err := seedApplications(ctx, queries); err != nil {
		return err
	}
	return normalizeTimestamps(ctx, dbtx)
}

// expect は生成された ID が ids.go の定数と一致することを検証する。
// ズレは「構築順と定数の不整合」なので、その場で止めて make e2e-dump を失敗させる
func expect(got, want int64, what string) error {
	if got != want {
		return fmt.Errorf("e2efixture: %s の ID が %d になった（ids.go の期待値は %d）。構築順か定数を見直すこと", what, got, want)
	}
	return nil
}

func seedUsers(ctx context.Context, queries *db.Queries) error {
	users := []struct {
		email  string
		role   string
		wantID int64
	}{
		{CompanyA.Email, "company", CompanyA.UserID},
		{CompanyB.Email, "company", CompanyB.UserID},
		{TalentA.Email, "talent", TalentA.UserID},
		{TalentB.Email, "talent", TalentB.UserID},
	}
	for _, u := range users {
		params := factories.CreateUserParams(factories.WithEmail(u.email), factories.WithRole(u.role))
		params.PasswordHash = passwordHash
		created, err := queries.CreateUser(ctx, params)
		if err != nil {
			return fmt.Errorf("e2efixture: ユーザー %s の作成に失敗: %w", u.email, err)
		}
		if err := expect(created.ID, u.wantID, "user "+u.email); err != nil {
			return err
		}
	}

	companyA, err := queries.CreateCompany(ctx, factories.CreateCompanyParams(CompanyA.UserID, factories.WithCompanyName("株式会社エー")))
	if err != nil {
		return fmt.Errorf("e2efixture: CompanyA プロフィール作成に失敗: %w", err)
	}
	if err := expect(companyA.ID, CompanyA.CompanyID, "company A"); err != nil {
		return err
	}
	companyB, err := queries.CreateCompany(ctx, factories.CreateCompanyParams(CompanyB.UserID, factories.WithCompanyName("株式会社ビー")))
	if err != nil {
		return fmt.Errorf("e2efixture: CompanyB プロフィール作成に失敗: %w", err)
	}
	if err := expect(companyB.ID, CompanyB.CompanyID, "company B"); err != nil {
		return err
	}

	talentA, err := queries.CreateTalent(ctx, factories.CreateTalentParams(TalentA.UserID, factories.WithDisplayName("人材エー")))
	if err != nil {
		return fmt.Errorf("e2efixture: TalentA プロフィール作成に失敗: %w", err)
	}
	if err := expect(talentA.ID, TalentA.TalentID, "talent A"); err != nil {
		return err
	}
	talentB, err := queries.CreateTalent(ctx, factories.CreateTalentParams(TalentB.UserID, factories.WithDisplayName("人材ビー")))
	if err != nil {
		return fmt.Errorf("e2efixture: TalentB プロフィール作成に失敗: %w", err)
	}
	return expect(talentB.ID, TalentB.TalentID, "talent B")
}

func seedProjects(ctx context.Context, queries *db.Queries) error {
	uc := companyuc.NewProject(queries)

	projects := []struct {
		ownerUserID int64
		title       string
		status      domain.ProjectStatus // draft のままなら空
		wantID      int64
	}{
		{CompanyA.UserID, "下書きの案件（A社）", "", Projects.ADraft},
		{CompanyA.UserID, "公開中の案件１（A社）", domain.ProjectPublished, Projects.APublished},
		{CompanyA.UserID, "公開中の案件２（A社）", domain.ProjectPublished, Projects.APublished2},
		{CompanyA.UserID, "募集終了した案件（A社）", domain.ProjectClosed, Projects.AClosed},
		{CompanyB.UserID, "公開中の案件（B社）", domain.ProjectPublished, Projects.BPublished},
	}
	for _, p := range projects {
		params := factories.CreateProjectParams()
		params.Title = p.title
		created, err := uc.Create(ctx, p.ownerUserID, params)
		if err != nil {
			return fmt.Errorf("e2efixture: 案件 %q の作成に失敗: %w", p.title, err)
		}
		if err := expect(created.ID, p.wantID, "project "+p.title); err != nil {
			return err
		}
		// closed へは遷移表どおり draft → published → closed の順で進める
		if p.status == domain.ProjectPublished || p.status == domain.ProjectClosed {
			if _, err := uc.ChangeStatus(ctx, p.ownerUserID, created.ID, domain.ProjectPublished); err != nil {
				return fmt.Errorf("e2efixture: 案件 %q の公開に失敗: %w", p.title, err)
			}
		}
		if p.status == domain.ProjectClosed {
			if _, err := uc.ChangeStatus(ctx, p.ownerUserID, created.ID, domain.ProjectClosed); err != nil {
				return fmt.Errorf("e2efixture: 案件 %q のクローズに失敗: %w", p.title, err)
			}
		}
	}
	return nil
}

// seedApplications は応募の全6状態を、遷移表（talent の Apply/Withdraw/Accept/Decline・
// company の Offer/Reject）だけを使って作る
func seedApplications(ctx context.Context, queries *db.Queries) error {
	talentApp := talentuc.NewApplication(queries)
	companyApp := companyuc.NewApplication(queries)

	apply := func(talentUserID, projectID int64, message string, wantID int64) (int64, error) {
		row, err := talentApp.Apply(ctx, talentUserID, projectID, message)
		if err != nil {
			return 0, fmt.Errorf("e2efixture: 応募（%s）の作成に失敗: %w", message, err)
		}
		return row.ID, expect(row.ID, wantID, "application "+message)
	}

	// 1. applied: TalentA → 公開案件1（応募したまま）
	if _, err := apply(TalentA.UserID, Projects.APublished, "応募中の基準応募です", Applications.Applied); err != nil {
		return err
	}

	// 2. offered: TalentB → 公開案件1 → A社がオファー
	id, err := apply(TalentB.UserID, Projects.APublished, "オファー済みの基準応募です", Applications.Offered)
	if err != nil {
		return err
	}
	if _, err := companyApp.Offer(ctx, CompanyA.UserID, id); err != nil {
		return fmt.Errorf("e2efixture: offered の遷移に失敗: %w", err)
	}

	// 3. accepted: TalentA → 公開案件2 → オファー → 承諾（ダブルオプトイン成立）
	id, err = apply(TalentA.UserID, Projects.APublished2, "承諾済みの基準応募です", Applications.Accepted)
	if err != nil {
		return err
	}
	if _, err := companyApp.Offer(ctx, CompanyA.UserID, id); err != nil {
		return fmt.Errorf("e2efixture: accepted のオファーに失敗: %w", err)
	}
	if _, err := talentApp.Accept(ctx, TalentA.UserID, id); err != nil {
		return fmt.Errorf("e2efixture: accepted の承諾に失敗: %w", err)
	}

	// 4. rejected: TalentB → 公開案件2 → 不採用
	id, err = apply(TalentB.UserID, Projects.APublished2, "不採用の基準応募です", Applications.Rejected)
	if err != nil {
		return err
	}
	if _, err := companyApp.Reject(ctx, CompanyA.UserID, id); err != nil {
		return fmt.Errorf("e2efixture: rejected の遷移に失敗: %w", err)
	}

	// 5. withdrawn: TalentA → B社の公開案件 → 取り下げ
	id, err = apply(TalentA.UserID, Projects.BPublished, "取り下げた基準応募です", Applications.Withdrawn)
	if err != nil {
		return err
	}
	if _, err := talentApp.Withdraw(ctx, TalentA.UserID, id); err != nil {
		return fmt.Errorf("e2efixture: withdrawn の遷移に失敗: %w", err)
	}

	// 6. declined: TalentB → B社の公開案件 → B社がオファー → 辞退
	id, err = apply(TalentB.UserID, Projects.BPublished, "辞退した基準応募です", Applications.Declined)
	if err != nil {
		return err
	}
	if _, err := companyApp.Offer(ctx, CompanyB.UserID, id); err != nil {
		return fmt.Errorf("e2efixture: declined のオファーに失敗: %w", err)
	}
	if _, err := talentApp.Decline(ctx, TalentB.UserID, id); err != nil {
		return fmt.Errorf("e2efixture: declined の辞退に失敗: %w", err)
	}
	return nil
}

// normalizeTimestamps は now() 由来の時刻を決定的な値に揃える。
// これをしないと seed のたびに dump.sql が変わり、再現性（CI 差分チェックの前提）が壊れる。
// 値は「基準時刻 + id 分」で行ごとにずらし、挿入順と時刻順の整合を保つ
func normalizeTimestamps(ctx context.Context, dbtx db.DBTX) error {
	stmts := []string{
		`UPDATE users SET created_at = '2026-01-01T00:00:00Z'::timestamptz + make_interval(mins => id::int)`,
		`UPDATE companies SET created_at = '2026-01-01T00:00:00Z'::timestamptz + make_interval(mins => id::int)`,
		`UPDATE talents SET created_at = '2026-01-01T00:00:00Z'::timestamptz + make_interval(mins => id::int)`,
		`UPDATE projects SET created_at = '2026-01-01T01:00:00Z'::timestamptz + make_interval(mins => id::int)`,
		`UPDATE applications SET created_at = '2026-01-01T02:00:00Z'::timestamptz + make_interval(mins => id::int)`,
		`UPDATE applications SET company_acted_at = created_at + interval '30 seconds' WHERE company_acted_at IS NOT NULL`,
		`UPDATE applications SET talent_acted_at = created_at + interval '60 seconds' WHERE talent_acted_at IS NOT NULL`,
	}
	for _, s := range stmts {
		if _, err := dbtx.Exec(ctx, s); err != nil {
			return fmt.Errorf("e2efixture: 時刻の正規化に失敗: %w", err)
		}
	}
	return nil
}
