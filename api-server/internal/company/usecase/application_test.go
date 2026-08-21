package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/domain"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/factories"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/helpers"
)

// setupPublishedProject は company を作り、公開済みの案件を1件用意する
func setupPublishedProject(t *testing.T, queries *db.Queries) (userID, projectID int64) {
	t.Helper()
	ctx := context.Background()
	uc := usecase.NewProject(queries)
	userID, _ = setupCompany(t, queries)
	project, err := uc.Create(ctx, userID, factories.CreateProjectParams())
	if err != nil {
		t.Fatalf("案件作成に失敗: %v", err)
	}
	if _, err := uc.ChangeStatus(ctx, userID, project.ID, domain.ProjectPublished); err != nil {
		t.Fatalf("公開に失敗: %v", err)
	}
	return userID, project.ID
}

// applyAsNewTalent は新しい talent を作って応募させ、応募 ID を返す
// （talent 側の usecase は境界ルールで import できないため queries を直接使う）
func applyAsNewTalent(t *testing.T, queries *db.Queries, projectID int64, message string) int64 {
	t.Helper()
	ctx := context.Background()
	user, err := queries.CreateUser(ctx, factories.CreateUserParams(factories.WithRole("talent")))
	if err != nil {
		t.Fatalf("ユーザー作成に失敗: %v", err)
	}
	tal, err := queries.CreateTalent(ctx, factories.CreateTalentParams(user.ID))
	if err != nil {
		t.Fatalf("人材プロフィール作成に失敗: %v", err)
	}
	app, err := queries.CreateApplication(ctx, db.CreateApplicationParams{
		TalentID:  tal.ID,
		Message:   message,
		ProjectID: projectID,
	})
	if err != nil {
		t.Fatalf("応募作成に失敗: %v", err)
	}
	return app.ID
}

// TestApplicationListForProject は自社案件の応募一覧を実DBで固定する。
//
// 目的: JOIN による応募者プロフィールの供給と、「他社の案件は404」「自社で応募ゼロは空リスト」の
// 区別（存在の有無を漏らさない）を保証する。
//
// 観点: 自社案件の応募が新しい順に表示名込みで返る／他社案件は ErrProjectNotFound／
// 自社案件で応募ゼロはエラーでなく0件。
func TestApplicationListForProject(t *testing.T) {
	ctx := context.Background()

	t.Run("自社案件の応募が応募者プロフィール込みで返る", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		userID, projectID := setupPublishedProject(t, queries)
		first := applyAsNewTalent(t, queries, projectID, "1人目です")
		second := applyAsNewTalent(t, queries, projectID, "2人目です")

		rows, err := uc.ListForProject(ctx, userID, projectID)
		if err != nil {
			t.Fatalf("一覧取得に失敗: %v", err)
		}
		if len(rows) != 2 {
			t.Fatalf("2件を期待したが %d 件", len(rows))
		}
		if rows[0].ID != second || rows[1].ID != first {
			t.Errorf("新しい順を期待したが: %d, %d", rows[0].ID, rows[1].ID)
		}
		if rows[0].TalentDisplayName == "" || len(rows[0].TalentSkills) == 0 {
			t.Errorf("応募者プロフィールが載っていない: %+v", rows[0])
		}
	})

	t.Run("他社の案件は404・自社で応募ゼロは空リスト（区別される）", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		_, projectID := setupPublishedProject(t, queries)
		otherUserID, _ := setupCompany(t, queries)

		// 他社（応募が付いている案件でも）→ 404
		applyAsNewTalent(t, queries, projectID, "")
		if _, err := uc.ListForProject(ctx, otherUserID, projectID); !errors.Is(err, usecase.ErrProjectNotFound) {
			t.Errorf("他社: ErrProjectNotFound を期待したが: %v", err)
		}

		// 自社の応募ゼロ案件 → エラーではなく0件
		ownUserID, ownProjectID := setupPublishedProject(t, queries)
		rows, err := uc.ListForProject(ctx, ownUserID, ownProjectID)
		if err != nil {
			t.Fatalf("応募ゼロの一覧が失敗: %v", err)
		}
		if len(rows) != 0 {
			t.Errorf("0件を期待したが %d 件", len(rows))
		}
	})
}

// TestApplicationOfferReject は選考の遷移を実DBで固定する。
//
// 目的: JOIN 越しの所有チェック（applications は company_id を持たない）・
// 遷移表（actor=company）の WHERE 反映・company_acted_at の同時記録を保証する。
//
// 観点: offer（applied→offered）／reject（applied→rejected）／二重オファー409
// （現在の状態つき）／他社の応募・不存在は同じ404。
func TestApplicationOfferReject(t *testing.T) {
	ctx := context.Background()

	t.Run("offer: applied → offered・company_acted_at が記録される", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		userID, projectID := setupPublishedProject(t, queries)
		appID := applyAsNewTalent(t, queries, projectID, "")

		row, err := uc.Offer(ctx, userID, appID, "")
		if err != nil {
			t.Fatalf("オファーに失敗: %v", err)
		}
		if row.Status != "offered" {
			t.Errorf("status: offered を期待したが %s", row.Status)
		}
		if !row.CompanyActedAt.Valid {
			t.Error("company_acted_at が記録されていない")
		}
		if row.TalentDisplayName == "" {
			t.Error("応募者プロフィールが載っていない")
		}
	})

	t.Run("reject: applied → rejected", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		userID, projectID := setupPublishedProject(t, queries)
		appID := applyAsNewTalent(t, queries, projectID, "")

		row, err := uc.Reject(ctx, userID, appID)
		if err != nil {
			t.Fatalf("不採用に失敗: %v", err)
		}
		if row.Status != "rejected" {
			t.Errorf("status: rejected を期待したが %s", row.Status)
		}
	})

	t.Run("二重オファーは ErrCannotChangeApplication（現在の状態つき409）", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		userID, projectID := setupPublishedProject(t, queries)
		appID := applyAsNewTalent(t, queries, projectID, "")

		if _, err := uc.Offer(ctx, userID, appID, ""); err != nil {
			t.Fatalf("1回目のオファーに失敗: %v", err)
		}
		_, err := uc.Offer(ctx, userID, appID, "")
		if !errors.Is(err, usecase.ErrCannotChangeApplication) {
			t.Fatalf("ErrCannotChangeApplication を期待したが: %v", err)
		}
		if !strings.Contains(err.Error(), "offered") {
			t.Errorf("エラーメッセージに現在の状態が含まれない: %v", err)
		}
	})

	t.Run("他社の応募と不存在は同じ ErrApplicationNotFound", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		_, projectID := setupPublishedProject(t, queries)
		otherUserID, _ := setupCompany(t, queries)
		appID := applyAsNewTalent(t, queries, projectID, "")

		if _, err := uc.Offer(ctx, otherUserID, appID, ""); !errors.Is(err, usecase.ErrApplicationNotFound) {
			t.Errorf("他社: ErrApplicationNotFound を期待したが: %v", err)
		}
		ownUserID, _ := setupPublishedProject(t, queries)
		if _, err := uc.Offer(ctx, ownUserID, 99999999, ""); !errors.Is(err, usecase.ErrApplicationNotFound) {
			t.Errorf("不存在: ErrApplicationNotFound を期待したが: %v", err)
		}
	})
}
