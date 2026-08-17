package usecase_test

import (
	"context"
	"slices"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/factories"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/helpers"
)

// TestProjectCreate は案件作成の業務フローを実DBで固定する。
//
// 目的: SQL・型の詰め替え・DB のデフォルト適用・制約まで含めて「本当に保存される」ことを
// 保証する。モックでは SQL の正しさは検証されない（設計プラン§10「テストは通るが動かない」対策）。
//
// 観点: 正常系（採番・status のデフォルト・入力値の往復）と、
// DB の CHECK 制約違反がエラーとして返ること（実DBテストでしか書けない異常系）。
func TestProjectCreate(t *testing.T) {
	t.Run("正常系: draft として保存され、入力値が往復する", func(t *testing.T) {
		uc := usecase.NewProject(helpers.NewTestQueries(t))

		params := factories.CreateProjectParams(factories.WithHourlyRate(3000, 5000))
		project, err := uc.Create(context.Background(), params)
		if err != nil {
			t.Fatalf("作成に失敗: %v", err)
		}

		if project.ID <= 0 {
			t.Error("id が採番されていない")
		}
		if project.Status != "draft" {
			t.Errorf("status: draft を期待したが %q", project.Status)
		}
		if project.CreatedAt.IsZero() {
			t.Error("created_at が設定されていない")
		}
		if project.Title != params.Title {
			t.Errorf("title: %q を期待したが %q", params.Title, project.Title)
		}
		if !slices.Equal(project.RequiredSkills, params.RequiredSkills) {
			t.Errorf("required_skills: %v を期待したが %v", params.RequiredSkills, project.RequiredSkills)
		}
		if project.HourlyRateMin == nil || *project.HourlyRateMin != 3000 {
			t.Errorf("hourly_rate_min: 3000 を期待したが %v", project.HourlyRateMin)
		}
	})

	t.Run("異常系: DB の CHECK 制約違反はエラーとして返る", func(t *testing.T) {
		uc := usecase.NewProject(helpers.NewTestQueries(t))

		// hours_per_week の許容範囲は 1〜60（マイグレーションの CHECK 制約）
		_, err := uc.Create(context.Background(), factories.CreateProjectParams(factories.WithHoursPerWeek(100)))
		if err == nil {
			t.Fatal("CHECK 制約違反がエラーにならなかった")
		}
	})
}
