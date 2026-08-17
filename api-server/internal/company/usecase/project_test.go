package usecase_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/usecase"
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
