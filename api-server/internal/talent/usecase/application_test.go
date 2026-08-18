package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/talent/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/factories"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/helpers"
)

// setupTalent は user と talent プロフィールを作り、userID を返す
func setupTalent(t *testing.T, queries *db.Queries) int64 {
	t.Helper()
	ctx := context.Background()
	user, err := queries.CreateUser(ctx, factories.CreateUserParams(factories.WithRole("talent")))
	if err != nil {
		t.Fatalf("ユーザー作成に失敗: %v", err)
	}
	if _, err := queries.CreateTalent(ctx, factories.CreateTalentParams(user.ID)); err != nil {
		t.Fatalf("人材プロフィール作成に失敗: %v", err)
	}
	return user.ID
}

// setOffered は応募を offered に進める（company の選考APIは #57 で実装するため生SQLで仕込む）
func setOffered(t *testing.T, tx pgx.Tx, applicationID int64) {
	t.Helper()
	tag, err := tx.Exec(context.Background(),
		"UPDATE applications SET status = 'offered' WHERE id = $1", applicationID)
	if err != nil || tag.RowsAffected() != 1 {
		t.Fatalf("offered への仕込みに失敗: %v", err)
	}
}

// TestApplicationApply は応募の作成を実DBで固定する。
//
// 目的: 「公開中の案件にのみ応募できる」（INSERT...SELECT の原子的検査）と
// 「二重応募は DB が拒否する」（UNIQUE 制約 → 409）を保証する。
//
// 観点: 正常系（applied で始まる・志望動機とタイトルが返る）／draft・closed・不存在への
// 応募が同一の404／二重応募409／プロフィール無し userID の拒否。
func TestApplicationApply(t *testing.T) {
	ctx := context.Background()

	t.Run("正常系: 公開中の案件に applied で応募できる", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		companyID := setupCompany(t, queries)
		project := createProject(t, queries, companyID, "published")
		userID := setupTalent(t, queries)

		row, err := uc.Apply(ctx, userID, project.ID, "ぜひやらせてください")
		if err != nil {
			t.Fatalf("応募に失敗: %v", err)
		}
		if row.Status != "applied" {
			t.Errorf("status: applied を期待したが %s", row.Status)
		}
		if row.Message != "ぜひやらせてください" {
			t.Errorf("message が保存されていない: %q", row.Message)
		}
		if row.ProjectTitle != project.Title {
			t.Errorf("project_title: %q を期待したが %q（JOIN が壊れている）", project.Title, row.ProjectTitle)
		}
	})

	t.Run("未公開（draft/closed）と不存在は同じ ErrProjectNotFound", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		companyID := setupCompany(t, queries)
		draft := createProject(t, queries, companyID, "draft")
		closed := createProject(t, queries, companyID, "closed")
		userID := setupTalent(t, queries)

		for name, id := range map[string]int64{
			"draft": draft.ID, "closed": closed.ID, "不存在": 99999999,
		} {
			if _, err := uc.Apply(ctx, userID, id, ""); !errors.Is(err, usecase.ErrProjectNotFound) {
				t.Errorf("%s: ErrProjectNotFound を期待したが: %v", name, err)
			}
		}
	})

	t.Run("二重応募は ErrAlreadyApplied（UNIQUE 制約）", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		companyID := setupCompany(t, queries)
		project := createProject(t, queries, companyID, "published")
		userID := setupTalent(t, queries)

		if _, err := uc.Apply(ctx, userID, project.ID, ""); err != nil {
			t.Fatalf("1回目の応募に失敗: %v", err)
		}
		if _, err := uc.Apply(ctx, userID, project.ID, ""); !errors.Is(err, usecase.ErrAlreadyApplied) {
			t.Errorf("ErrAlreadyApplied を期待したが: %v", err)
		}
	})

	t.Run("プロフィールの無い userID は拒否される", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)

		if _, err := uc.Apply(ctx, 99999999, 1, ""); !errors.Is(err, usecase.ErrAuthFailed) {
			t.Errorf("ErrAuthFailed を期待したが: %v", err)
		}
	})
}

// TestApplicationListMine は自分の応募一覧を実DBで固定する。
//
// 目的: WHERE talent_id による分離（他人の応募が混ざらない）と並び順・JOIN のタイトルを保証する。
// 観点: 2人の talent がそれぞれ応募し、自分の分だけが新しい順で返ること。
func TestApplicationListMine(t *testing.T) {
	ctx := context.Background()
	_, queries := helpers.NewTestTx(t)
	uc := usecase.NewApplication(queries)
	companyID := setupCompany(t, queries)
	first := createProject(t, queries, companyID, "published")
	second := createProject(t, queries, companyID, "published")
	me := setupTalent(t, queries)
	other := setupTalent(t, queries)

	if _, err := uc.Apply(ctx, me, first.ID, ""); err != nil {
		t.Fatalf("応募に失敗: %v", err)
	}
	if _, err := uc.Apply(ctx, me, second.ID, ""); err != nil {
		t.Fatalf("応募に失敗: %v", err)
	}
	if _, err := uc.Apply(ctx, other, first.ID, ""); err != nil {
		t.Fatalf("他人の応募に失敗: %v", err)
	}

	rows, err := uc.ListMine(ctx, me)
	if err != nil {
		t.Fatalf("一覧取得に失敗: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("自分の2件を期待したが %d 件（他人の応募が混ざっていないか）", len(rows))
	}
	if rows[0].ProjectID != second.ID || rows[1].ProjectID != first.ID {
		t.Errorf("新しい順を期待したが: %d, %d", rows[0].ProjectID, rows[1].ProjectID)
	}
	if rows[0].ProjectTitle == "" {
		t.Error("project_title が空（JOIN が壊れている）")
	}
}

// TestApplicationWithdraw は取り下げを実DBで固定する。
//
// 目的: 遷移表（talent は applied / offered から withdrawn へ）が SQL の WHERE に正しく
// 反映されること・所有チェック・決着済みの巻き戻し禁止を保証する。
//
// 観点: applied から／offered から（生SQLで仕込み）／決着済み（withdrawn）は409で
// 現在の状態がメッセージに含まれる／他人の応募と不存在は同じ404。
func TestApplicationWithdraw(t *testing.T) {
	ctx := context.Background()

	t.Run("applied から取り下げできる（talent_acted_at も記録される）", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		companyID := setupCompany(t, queries)
		project := createProject(t, queries, companyID, "published")
		userID := setupTalent(t, queries)
		app, err := uc.Apply(ctx, userID, project.ID, "")
		if err != nil {
			t.Fatalf("応募に失敗: %v", err)
		}

		row, err := uc.Withdraw(ctx, userID, app.ID)
		if err != nil {
			t.Fatalf("取り下げに失敗: %v", err)
		}
		if row.Status != "withdrawn" {
			t.Errorf("status: withdrawn を期待したが %s", row.Status)
		}
		if !row.TalentActedAt.Valid {
			t.Error("talent_acted_at が記録されていない")
		}
	})

	t.Run("offered からも取り下げできる", func(t *testing.T) {
		tx, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		companyID := setupCompany(t, queries)
		project := createProject(t, queries, companyID, "published")
		userID := setupTalent(t, queries)
		app, err := uc.Apply(ctx, userID, project.ID, "")
		if err != nil {
			t.Fatalf("応募に失敗: %v", err)
		}
		setOffered(t, tx, app.ID)

		row, err := uc.Withdraw(ctx, userID, app.ID)
		if err != nil {
			t.Fatalf("取り下げに失敗: %v", err)
		}
		if row.Status != "withdrawn" {
			t.Errorf("status: withdrawn を期待したが %s", row.Status)
		}
	})

	t.Run("決着済み（withdrawn）は ErrCannotChangeApplication（現在の状態つき409）", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		companyID := setupCompany(t, queries)
		project := createProject(t, queries, companyID, "published")
		userID := setupTalent(t, queries)
		app, err := uc.Apply(ctx, userID, project.ID, "")
		if err != nil {
			t.Fatalf("応募に失敗: %v", err)
		}
		if _, err := uc.Withdraw(ctx, userID, app.ID); err != nil {
			t.Fatalf("1回目の取り下げに失敗: %v", err)
		}

		_, err = uc.Withdraw(ctx, userID, app.ID)
		if !errors.Is(err, usecase.ErrCannotChangeApplication) {
			t.Fatalf("ErrCannotChangeApplication を期待したが: %v", err)
		}
		if !strings.Contains(err.Error(), "withdrawn") {
			t.Errorf("エラーメッセージに現在の状態が含まれない: %v", err)
		}
	})

	t.Run("他人の応募と不存在は同じ ErrApplicationNotFound", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		companyID := setupCompany(t, queries)
		project := createProject(t, queries, companyID, "published")
		owner := setupTalent(t, queries)
		other := setupTalent(t, queries)
		app, err := uc.Apply(ctx, owner, project.ID, "")
		if err != nil {
			t.Fatalf("応募に失敗: %v", err)
		}

		if _, err := uc.Withdraw(ctx, other, app.ID); !errors.Is(err, usecase.ErrApplicationNotFound) {
			t.Errorf("他人: ErrApplicationNotFound を期待したが: %v", err)
		}
		if _, err := uc.Withdraw(ctx, owner, 99999999); !errors.Is(err, usecase.ErrApplicationNotFound) {
			t.Errorf("不存在: ErrApplicationNotFound を期待したが: %v", err)
		}
	})
}

// TestApplicationAcceptDecline は承諾・辞退を実DBで固定する。
//
// 目的: 遷移表（talent の accept / decline は offered からのみ）が WHERE に反映され、
// applied のまま承諾できない＝ダブルオプトインの片翼だけでは成立しないことを保証する。
//
// 観点: offered からの accept / decline（talent_acted_at 記録）／applied のままの
// accept は409（現在の状態つき）。
func TestApplicationAcceptDecline(t *testing.T) {
	ctx := context.Background()

	t.Run("accept: offered → accepted", func(t *testing.T) {
		tx, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		companyID := setupCompany(t, queries)
		project := createProject(t, queries, companyID, "published")
		userID := setupTalent(t, queries)
		app, err := uc.Apply(ctx, userID, project.ID, "")
		if err != nil {
			t.Fatalf("応募に失敗: %v", err)
		}
		setOffered(t, tx, app.ID)

		row, err := uc.Accept(ctx, userID, app.ID)
		if err != nil {
			t.Fatalf("承諾に失敗: %v", err)
		}
		if row.Status != "accepted" {
			t.Errorf("status: accepted を期待したが %s", row.Status)
		}
		if !row.TalentActedAt.Valid {
			t.Error("talent_acted_at が記録されていない")
		}
	})

	t.Run("decline: offered → declined", func(t *testing.T) {
		tx, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		companyID := setupCompany(t, queries)
		project := createProject(t, queries, companyID, "published")
		userID := setupTalent(t, queries)
		app, err := uc.Apply(ctx, userID, project.ID, "")
		if err != nil {
			t.Fatalf("応募に失敗: %v", err)
		}
		setOffered(t, tx, app.ID)

		row, err := uc.Decline(ctx, userID, app.ID)
		if err != nil {
			t.Fatalf("辞退に失敗: %v", err)
		}
		if row.Status != "declined" {
			t.Errorf("status: declined を期待したが %s", row.Status)
		}
	})

	t.Run("applied のまま accept は409（片方の意思では成立しない）", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewApplication(queries)
		companyID := setupCompany(t, queries)
		project := createProject(t, queries, companyID, "published")
		userID := setupTalent(t, queries)
		app, err := uc.Apply(ctx, userID, project.ID, "")
		if err != nil {
			t.Fatalf("応募に失敗: %v", err)
		}

		_, err = uc.Accept(ctx, userID, app.ID)
		if !errors.Is(err, usecase.ErrCannotChangeApplication) {
			t.Fatalf("ErrCannotChangeApplication を期待したが: %v", err)
		}
		if !strings.Contains(err.Error(), "applied") {
			t.Errorf("エラーメッセージに現在の状態が含まれない: %v", err)
		}
	})
}

// TestApplicationDoubleOptInFlow は状態機械の一巡を実DBで固定する統合テスト。
//
// 目的: applied →(company)→ offered →(talent)→ accepted という「両者の合意の連鎖」が
// 唯一の成立経路であること、accepted が終端（以後どの操作も409）であることを保証する。
//
// 観点: 各段階での不正操作の拒否と正規経路の成立、終端での全操作拒否。
func TestApplicationDoubleOptInFlow(t *testing.T) {
	ctx := context.Background()
	tx, queries := helpers.NewTestTx(t)
	uc := usecase.NewApplication(queries)
	companyID := setupCompany(t, queries)
	project := createProject(t, queries, companyID, "published")
	userID := setupTalent(t, queries)

	// 応募（applied）
	app, err := uc.Apply(ctx, userID, project.ID, "一巡テスト")
	if err != nil {
		t.Fatalf("応募に失敗: %v", err)
	}

	// applied のままでは accept も decline もできない
	if _, err := uc.Accept(ctx, userID, app.ID); !errors.Is(err, usecase.ErrCannotChangeApplication) {
		t.Errorf("applied での accept: 409 を期待したが: %v", err)
	}
	if _, err := uc.Decline(ctx, userID, app.ID); !errors.Is(err, usecase.ErrCannotChangeApplication) {
		t.Errorf("applied での decline: 409 を期待したが: %v", err)
	}

	// company がオファー（#57 の usecase は境界ルールで使えないため生SQL）
	setOffered(t, tx, app.ID)

	// talent が承諾 → ダブルオプトイン成立
	row, err := uc.Accept(ctx, userID, app.ID)
	if err != nil {
		t.Fatalf("承諾に失敗: %v", err)
	}
	if row.Status != "accepted" {
		t.Fatalf("status: accepted を期待したが %s", row.Status)
	}

	// accepted は終端: talent のどの操作も409
	for name, op := range map[string]func(context.Context, int64, int64) (any, error){
		"withdraw": func(ctx context.Context, u, id int64) (any, error) { return uc.Withdraw(ctx, u, id) },
		"accept":   func(ctx context.Context, u, id int64) (any, error) { return uc.Accept(ctx, u, id) },
		"decline":  func(ctx context.Context, u, id int64) (any, error) { return uc.Decline(ctx, u, id) },
	} {
		if _, err := op(ctx, userID, app.ID); !errors.Is(err, usecase.ErrCannotChangeApplication) {
			t.Errorf("終端での %s: 409 を期待したが: %v", name, err)
		}
	}
}
