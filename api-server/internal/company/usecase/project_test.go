package usecase_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/domain"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/factories"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/helpers"
)

// TestProjectCreate は案件作成の業務フローを実DBで固定する。
//
// 目的: SQL・型の詰め替え・DB のデフォルト適用・制約・**所有者の決定**まで含めて
// 「本当に正しく保存される」ことを保証する（設計プラン§10）。
//
// 観点: 正常系（採番・status のデフォルト・**company_id がトークン由来の userID から解決される**）／
// プロフィールの無い userID の拒否（IDOR 対策の入口）／nil スライスの正規化／
// DB の CHECK 制約違反（実DBテストでしか書けない異常系）。
func TestProjectCreate(t *testing.T) {
	// user と company プロフィールを作り、その userID と companyID を返す
	setup := func(t *testing.T, queries *db.Queries) (int64, int64) {
		t.Helper()
		ctx := context.Background()
		user, err := queries.CreateUser(ctx, factories.CreateUserParams())
		if err != nil {
			t.Fatalf("ユーザー作成に失敗: %v", err)
		}
		comp, err := queries.CreateCompany(ctx, factories.CreateCompanyParams(user.ID))
		if err != nil {
			t.Fatalf("企業プロフィール作成に失敗: %v", err)
		}
		return user.ID, comp.ID
	}

	t.Run("正常系: draft として保存され、所有者がトークン由来で決まる", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewProject(queries)
		userID, companyID := setup(t, queries)

		params := factories.CreateProjectParams(factories.WithHourlyRate(3000, 5000))
		project, err := uc.Create(context.Background(), userID, params)
		if err != nil {
			t.Fatalf("作成に失敗: %v", err)
		}

		if project.CompanyID != companyID {
			t.Errorf("company_id: %d を期待したが %d（トークン由来の解決が壊れている）", companyID, project.CompanyID)
		}
		if project.ID <= 0 {
			t.Error("id が採番されていない")
		}
		if project.Status != "draft" {
			t.Errorf("status: draft を期待したが %q", project.Status)
		}
		if !slices.Equal(project.RequiredSkills, params.RequiredSkills) {
			t.Errorf("required_skills: %v を期待したが %v", params.RequiredSkills, project.RequiredSkills)
		}
	})

	t.Run("プロフィールの無い userID は拒否される", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewProject(queries)

		_, err := uc.Create(context.Background(), 99999999, factories.CreateProjectParams())
		if !errors.Is(err, usecase.ErrAuthFailed) {
			t.Errorf("ErrAuthFailed を期待したが: %v", err)
		}
	})

	t.Run("nil の required_skills は空配列として保存される", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewProject(queries)
		userID, _ := setup(t, queries)

		params := factories.CreateProjectParams()
		params.RequiredSkills = nil // NOT NULL 制約に対する正規化（#30 と同型）の検証

		project, err := uc.Create(context.Background(), userID, params)
		if err != nil {
			t.Fatalf("作成に失敗: %v", err)
		}
		if project.RequiredSkills == nil || len(project.RequiredSkills) != 0 {
			t.Errorf("空配列を期待したが %v", project.RequiredSkills)
		}
	})

	t.Run("異常系: DB の CHECK 制約違反はエラーとして返る", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewProject(queries)
		userID, _ := setup(t, queries)

		// hours_per_week の許容範囲は 1〜60（マイグレーションの CHECK 制約）
		_, err := uc.Create(context.Background(), userID, factories.CreateProjectParams(factories.WithHoursPerWeek(100)))
		if err == nil {
			t.Fatal("CHECK 制約違反がエラーにならなかった")
		}
	})
}

// TestProjectChangeStatus は掲載状態の遷移を実DBで固定する。
//
// 目的: 遷移表（domain）・所有チェック（WHERE company_id）・条件付き UPDATE が
// 揃って初めて安全な状態遷移になることを保証する。
//
// 観点: 掲載フローの一巡（公開→非公開→公開→終了→再募集）／不許可遷移が
// TransitionError（可能な遷移先つき）になること／他社・不存在の案件は
// 区別されず ErrProjectNotFound になること（存在の有無を漏らさない）。
func TestProjectChangeStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("掲載フローの一巡: draft→published→draft→published→closed→published", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewProject(queries)
		userID, _ := setupCompany(t, queries)
		project, err := uc.Create(ctx, userID, factories.CreateProjectParams())
		if err != nil {
			t.Fatalf("作成に失敗: %v", err)
		}

		steps := []domain.ProjectStatus{
			domain.ProjectPublished, // 公開
			domain.ProjectDraft,     // 非公開
			domain.ProjectPublished, // 再公開
			domain.ProjectClosed,    // 募集終了
			domain.ProjectPublished, // 再募集
		}
		for _, to := range steps {
			updated, err := uc.ChangeStatus(ctx, userID, project.ID, to)
			if err != nil {
				t.Fatalf("%s への遷移に失敗: %v", to, err)
			}
			if updated.Status != string(to) {
				t.Fatalf("status: %s を期待したが %s", to, updated.Status)
			}
		}
	})

	t.Run("不許可の遷移は TransitionError（可能な遷移先つき）", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewProject(queries)
		userID, _ := setupCompany(t, queries)
		project, err := uc.Create(ctx, userID, factories.CreateProjectParams())
		if err != nil {
			t.Fatalf("作成に失敗: %v", err)
		}

		// draft → closed は表に無い
		_, err = uc.ChangeStatus(ctx, userID, project.ID, domain.ProjectClosed)
		var transitionErr *usecase.TransitionError
		if !errors.As(err, &transitionErr) {
			t.Fatalf("TransitionError を期待したが: %v", err)
		}
		if !strings.Contains(transitionErr.Error(), "published") {
			t.Errorf("エラーメッセージに可能な遷移先が含まれない: %v", transitionErr)
		}
	})

	t.Run("他社の案件と不存在の案件は同じ ErrProjectNotFound", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewProject(queries)
		ownerID, _ := setupCompany(t, queries)
		otherID, _ := setupCompany(t, queries)
		project, err := uc.Create(ctx, ownerID, factories.CreateProjectParams())
		if err != nil {
			t.Fatalf("作成に失敗: %v", err)
		}

		// 他社（otherID）が所有者の案件を publish しようとする
		_, err = uc.ChangeStatus(ctx, otherID, project.ID, domain.ProjectPublished)
		if !errors.Is(err, usecase.ErrProjectNotFound) {
			t.Errorf("他社: ErrProjectNotFound を期待したが: %v", err)
		}

		_, err = uc.ChangeStatus(ctx, ownerID, 99999999, domain.ProjectPublished)
		if !errors.Is(err, usecase.ErrProjectNotFound) {
			t.Errorf("不存在: ErrProjectNotFound を期待したが: %v", err)
		}
	})
}

// setupCompany は user と company プロフィールを作り、userID と companyID を返す共通ヘルパー
func setupCompany(t *testing.T, queries *db.Queries) (int64, int64) {
	t.Helper()
	ctx := context.Background()
	user, err := queries.CreateUser(ctx, factories.CreateUserParams())
	if err != nil {
		t.Fatalf("ユーザー作成に失敗: %v", err)
	}
	comp, err := queries.CreateCompany(ctx, factories.CreateCompanyParams(user.ID))
	if err != nil {
		t.Fatalf("企業プロフィール作成に失敗: %v", err)
	}
	return user.ID, comp.ID
}

// TestProjectListMine は自社案件の一覧を実DBで固定する。
//
// 目的: 所有チェックが WHERE にあること（他社の案件は一覧に存在しない）と並び順を保証する。
// 観点: 2社を作り、自社の分だけが新しい順で返ること。
func TestProjectListMine(t *testing.T) {
	ctx := context.Background()
	_, queries := helpers.NewTestTx(t)
	uc := usecase.NewProject(queries)

	ownerID, _ := setupCompany(t, queries)
	otherID, _ := setupCompany(t, queries)

	first, err := uc.Create(ctx, ownerID, factories.CreateProjectParams())
	if err != nil {
		t.Fatalf("作成に失敗: %v", err)
	}
	second, err := uc.Create(ctx, ownerID, factories.CreateProjectParams())
	if err != nil {
		t.Fatalf("作成に失敗: %v", err)
	}
	if _, err := uc.Create(ctx, otherID, factories.CreateProjectParams()); err != nil {
		t.Fatalf("他社の作成に失敗: %v", err)
	}

	projects, err := uc.ListMine(ctx, ownerID)
	if err != nil {
		t.Fatalf("一覧取得に失敗: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("自社の2件を期待したが %d 件（他社の案件が混ざっていないか）", len(projects))
	}
	if projects[0].ID != second.ID || projects[1].ID != first.ID {
		t.Errorf("新しい順を期待したが: %d, %d", projects[0].ID, projects[1].ID)
	}
}

// TestProjectUpdate は案件編集を実DBで固定する。
//
// 目的: 「編集で公開状態が変わらない」（SET 句に status が無い）ことと所有チェックを保証する。
// 観点: published の案件を編集しても published のまま／値は更新される／他社は404相当。
func TestProjectUpdate(t *testing.T) {
	ctx := context.Background()

	t.Run("値は更新され、公開状態は変わらない", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewProject(queries)
		userID, _ := setupCompany(t, queries)

		project, err := uc.Create(ctx, userID, factories.CreateProjectParams())
		if err != nil {
			t.Fatalf("作成に失敗: %v", err)
		}
		if _, err := uc.ChangeStatus(ctx, userID, project.ID, domain.ProjectPublished); err != nil {
			t.Fatalf("公開に失敗: %v", err)
		}

		updated, err := uc.Update(ctx, userID, project.ID, db.UpdateProjectParams{
			Title:        "更新後のタイトル",
			Description:  "更新後の詳細",
			HoursPerWeek: 20,
			RemoteOk:     false,
		})
		if err != nil {
			t.Fatalf("更新に失敗: %v", err)
		}
		if updated.Title != "更新後のタイトル" || updated.HoursPerWeek != 20 {
			t.Errorf("値が更新されていない: %+v", updated)
		}
		if updated.Status != string(domain.ProjectPublished) {
			t.Errorf("編集で status が変わった: %s（SET句にstatusが混入していないか）", updated.Status)
		}
		if updated.RequiredSkills == nil {
			t.Error("nil skills が正規化されていない")
		}
	})

	t.Run("他社の案件は更新できない（ErrProjectNotFound）", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewProject(queries)
		ownerID, _ := setupCompany(t, queries)
		otherID, _ := setupCompany(t, queries)

		project, err := uc.Create(ctx, ownerID, factories.CreateProjectParams())
		if err != nil {
			t.Fatalf("作成に失敗: %v", err)
		}

		_, err = uc.Update(ctx, otherID, project.ID, db.UpdateProjectParams{
			Title: "乗っ取り", Description: "x", HoursPerWeek: 10, RemoteOk: true,
		})
		if !errors.Is(err, usecase.ErrProjectNotFound) {
			t.Errorf("ErrProjectNotFound を期待したが: %v", err)
		}
	})
}
