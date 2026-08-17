package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/talent/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/factories"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/helpers"
)

// setupCompany は user と company プロフィールを作り companyID を返す。
// talent 側のテストだが、公開案件を用意するには発注側（company）のデータが必要
func setupCompany(t *testing.T, queries *db.Queries) int64 {
	t.Helper()
	ctx := context.Background()
	user, err := queries.CreateUser(ctx, factories.CreateUserParams(factories.WithRole("company")))
	if err != nil {
		t.Fatalf("ユーザー作成に失敗: %v", err)
	}
	comp, err := queries.CreateCompany(ctx, factories.CreateCompanyParams(user.ID))
	if err != nil {
		t.Fatalf("企業プロフィール作成に失敗: %v", err)
	}
	return comp.ID
}

// createProject は指定 status の案件を1件用意する。
// company/usecase は境界ルール（depguard）で import できないため、queries を直接使う
func createProject(t *testing.T, queries *db.Queries, companyID int64, status string, opts ...factories.ProjectOption) db.Project {
	t.Helper()
	ctx := context.Background()

	params := factories.CreateProjectParams(opts...)
	params.CompanyID = companyID
	project, err := queries.CreateProject(ctx, params)
	if err != nil {
		t.Fatalf("案件作成に失敗: %v", err)
	}
	if status == "draft" {
		return project
	}

	// draft → published →（必要なら）closed と遷移表どおりに進める
	project = changeStatus(t, queries, project, "draft", "published")
	if status == "closed" {
		project = changeStatus(t, queries, project, "published", "closed")
	}
	return project
}

func changeStatus(t *testing.T, queries *db.Queries, project db.Project, from, to string) db.Project {
	t.Helper()
	updated, err := queries.UpdateProjectStatus(context.Background(), db.UpdateProjectStatusParams{
		ID:         project.ID,
		CompanyID:  project.CompanyID,
		FromStatus: from,
		ToStatus:   to,
	})
	if err != nil {
		t.Fatalf("状態遷移（%s→%s）に失敗: %v", from, to, err)
	}
	return updated
}

// TestListPublishedSeekStability は seek 法の核心＝挿入ズレ耐性を実DBで固定する。
//
// 目的: 1ページ目を読んだ後に新規公開が挟まっても、2ページ目が重複も欠落もしないことを保証する
// （OFFSET 法だと全件が1つ後ろにズレて、1ページ目の末尾が2ページ目に再登場する）。
//
// 観点: cursor（id <= next_cursor）で読む2ページ目が「1ページ目時点の続き」を正確に返すこと／
// 新規公開分は次回の1ページ目に現れること。
func TestListPublishedSeekStability(t *testing.T) {
	ctx := context.Background()
	_, queries := helpers.NewTestTx(t)
	uc := usecase.NewProject(queries)
	companyID := setupCompany(t, queries)

	// 公開案件を5件用意（id 昇順に p[0]..p[4]。一覧は id 降順なので p[4] が先頭）
	published := make([]db.Project, 5)
	for i := range published {
		published[i] = createProject(t, queries, companyID, "published")
	}

	// 1ページ目（limit=2）: 新しい順に p[4], p[3] が返り、next_cursor は p[2].ID
	page1, err := uc.ListPublished(ctx, usecase.ListPublishedParams{Limit: new(int32(2))})
	if err != nil {
		t.Fatalf("1ページ目の取得に失敗: %v", err)
	}
	if len(page1.Projects) != 2 {
		t.Fatalf("1ページ目: 2件を期待したが %d 件", len(page1.Projects))
	}
	if page1.Projects[0].ID != published[4].ID || page1.Projects[1].ID != published[3].ID {
		t.Fatalf("1ページ目の並び: [%d, %d] を期待したが [%d, %d]",
			published[4].ID, published[3].ID, page1.Projects[0].ID, page1.Projects[1].ID)
	}
	if page1.NextCursor == nil || *page1.NextCursor != published[2].ID {
		t.Fatalf("next_cursor: %d を期待したが %v", published[2].ID, page1.NextCursor)
	}

	// ここで新規公開が割り込む（OFFSET 法ならこの1件が全ページをズラす）
	createProject(t, queries, companyID, "published")

	// 2ページ目: cursor 起点なので割り込みの影響を受けず p[2], p[1] が返る
	page2, err := uc.ListPublished(ctx, usecase.ListPublishedParams{
		Cursor: page1.NextCursor,
		Limit:  new(int32(2)),
	})
	if err != nil {
		t.Fatalf("2ページ目の取得に失敗: %v", err)
	}
	if len(page2.Projects) != 2 {
		t.Fatalf("2ページ目: 2件を期待したが %d 件", len(page2.Projects))
	}
	if page2.Projects[0].ID != published[2].ID || page2.Projects[1].ID != published[1].ID {
		t.Errorf("2ページ目がズレた: [%d, %d] を期待したが [%d, %d]（重複・欠落の疑い）",
			published[2].ID, published[1].ID, page2.Projects[0].ID, page2.Projects[1].ID)
	}
	if page2.NextCursor == nil || *page2.NextCursor != published[0].ID {
		t.Errorf("2ページ目の next_cursor: %d を期待したが %v", published[0].ID, page2.NextCursor)
	}
}

// TestListPublishedPageBoundary はページ境界（n+1 テクニックの分岐）を両側から固定する。
//
// 目的: 「ちょうど limit 件」で next_cursor が nil になり（余計な次ページ案内を出さない）、
// 「limit+1 件以上」で limit 件に切り詰めて next_cursor が付くことを保証する。
//
// 観点: 総件数 == limit（次ページなし）／総件数 == limit+1（ちょうど1件はみ出す）の境界2点。
func TestListPublishedPageBoundary(t *testing.T) {
	ctx := context.Background()

	t.Run("ちょうど limit 件なら next_cursor は nil", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewProject(queries)
		companyID := setupCompany(t, queries)
		for range 3 {
			createProject(t, queries, companyID, "published")
		}

		page, err := uc.ListPublished(ctx, usecase.ListPublishedParams{Limit: new(int32(3))})
		if err != nil {
			t.Fatalf("取得に失敗: %v", err)
		}
		if len(page.Projects) != 3 {
			t.Errorf("3件を期待したが %d 件", len(page.Projects))
		}
		if page.NextCursor != nil {
			t.Errorf("next_cursor は nil を期待したが %d（n+1 の判定が壊れている）", *page.NextCursor)
		}
	})

	t.Run("limit+1 件あれば limit 件に切り詰めて next_cursor が付く", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewProject(queries)
		companyID := setupCompany(t, queries)
		for range 4 {
			createProject(t, queries, companyID, "published")
		}

		page, err := uc.ListPublished(ctx, usecase.ListPublishedParams{Limit: new(int32(3))})
		if err != nil {
			t.Fatalf("取得に失敗: %v", err)
		}
		if len(page.Projects) != 3 {
			t.Errorf("3件に切り詰められるはずが %d 件（n+1 件目が漏れている）", len(page.Projects))
		}
		if page.NextCursor == nil {
			t.Error("next_cursor が付いていない")
		}
	})

	t.Run("limit 未指定は既定の20件・上限50超は50にクランプ", func(t *testing.T) {
		_, queries := helpers.NewTestTx(t)
		uc := usecase.NewProject(queries)
		companyID := setupCompany(t, queries)
		for range 51 {
			createProject(t, queries, companyID, "published")
		}

		byDefault, err := uc.ListPublished(ctx, usecase.ListPublishedParams{})
		if err != nil {
			t.Fatalf("取得に失敗: %v", err)
		}
		if len(byDefault.Projects) != 20 {
			t.Errorf("既定の20件を期待したが %d 件", len(byDefault.Projects))
		}

		clamped, err := uc.ListPublished(ctx, usecase.ListPublishedParams{Limit: new(int32(1000))})
		if err != nil {
			t.Fatalf("取得に失敗: %v", err)
		}
		if len(clamped.Projects) != 50 {
			t.Errorf("上限50件へのクランプを期待したが %d 件（巨大 limit が素通りしている）", len(clamped.Projects))
		}
	})
}

// TestListPublishedFilters は検索条件の合成を実DBで固定する。
//
// 目的: narg（NULL＝条件無効）と @>（配列の包含＝スキル AND 検索）の組み合わせが
// 意図どおりに絞り込むことを保証する。
//
// 観点: skills は AND（全部持つ案件だけ）／条件未指定は全件／remote_ok・min_hourly_rate の絞り込み。
func TestListPublishedFilters(t *testing.T) {
	ctx := context.Background()
	_, queries := helpers.NewTestTx(t)
	uc := usecase.NewProject(queries)
	companyID := setupCompany(t, queries)

	goPg := createProject(t, queries, companyID, "published") // skills: Go, PostgreSQL（factory 既定）

	goOnlyParams := factories.CreateProjectParams()
	goOnlyParams.RequiredSkills = []string{"Go"}
	goOnlyParams.CompanyID = companyID
	goOnly, err := queries.CreateProject(ctx, goOnlyParams)
	if err != nil {
		t.Fatalf("案件作成に失敗: %v", err)
	}
	changeStatus(t, queries, goOnly, "draft", "published")

	remote := createProject(t, queries, companyID, "published", factories.WithHourlyRate(5000, 8000))

	t.Run("skills は AND: 指定スキルを全部持つ案件だけ返る", func(t *testing.T) {
		page, err := uc.ListPublished(ctx, usecase.ListPublishedParams{
			Skills: []string{"Go", "PostgreSQL"},
		})
		if err != nil {
			t.Fatalf("取得に失敗: %v", err)
		}
		// goOnly（Go のみ）は除外され、既定スキルの2件（goPg, remote）が残る
		if len(page.Projects) != 2 {
			t.Fatalf("2件を期待したが %d 件", len(page.Projects))
		}
		for _, p := range page.Projects {
			if p.ID == goOnly.ID {
				t.Errorf("Go しか持たない案件が AND 検索に混入した（@> が OR になっていないか）")
			}
		}
	})

	t.Run("空の skills は絞り込みなし（全件）", func(t *testing.T) {
		page, err := uc.ListPublished(ctx, usecase.ListPublishedParams{Skills: []string{}})
		if err != nil {
			t.Fatalf("取得に失敗: %v", err)
		}
		if len(page.Projects) != 3 {
			t.Errorf("全3件を期待したが %d 件（空配列の正規化が壊れている）", len(page.Projects))
		}
	})

	t.Run("min_hourly_rate: 下限以上の案件だけ返る", func(t *testing.T) {
		page, err := uc.ListPublished(ctx, usecase.ListPublishedParams{
			MinHourlyRate: new(int32(4000)),
		})
		if err != nil {
			t.Fatalf("取得に失敗: %v", err)
		}
		if len(page.Projects) != 1 || page.Projects[0].ID != remote.ID {
			t.Errorf("時給5000の1件のみを期待したが %d 件", len(page.Projects))
		}
		_ = goPg
	})
}

// TestListPublishedOnlyPublished は「公開中だけが見える」ことを実DBで固定する。
//
// 目的: talent 一覧に draft（下書き）・closed（募集終了）が混入しないことを保証する
// （混入は他社の未公開情報の漏えい＝セキュリティ要件）。
//
// 観点: draft / published / closed を1件ずつ作り、published だけが返ること。
func TestListPublishedOnlyPublished(t *testing.T) {
	ctx := context.Background()
	_, queries := helpers.NewTestTx(t)
	uc := usecase.NewProject(queries)
	companyID := setupCompany(t, queries)

	createProject(t, queries, companyID, "draft")
	published := createProject(t, queries, companyID, "published")
	createProject(t, queries, companyID, "closed")

	page, err := uc.ListPublished(ctx, usecase.ListPublishedParams{})
	if err != nil {
		t.Fatalf("取得に失敗: %v", err)
	}
	if len(page.Projects) != 1 || page.Projects[0].ID != published.ID {
		t.Fatalf("published の1件のみを期待したが %d 件（未公開が漏れている）", len(page.Projects))
	}
}

// TestGetPublished は詳細取得の公開境界を実DBで固定する。
//
// 目的: 未公開（draft / closed）と不存在が同じ ErrProjectNotFound になることを保証する
// （id を総当たりしても未公開案件の存在自体が分からない）。
//
// 観点: published は取得できる／draft・closed・不存在 id はすべて同一エラー。
func TestGetPublished(t *testing.T) {
	ctx := context.Background()
	_, queries := helpers.NewTestTx(t)
	uc := usecase.NewProject(queries)
	companyID := setupCompany(t, queries)

	published := createProject(t, queries, companyID, "published")
	draft := createProject(t, queries, companyID, "draft")
	closed := createProject(t, queries, companyID, "closed")

	got, err := uc.GetPublished(ctx, published.ID)
	if err != nil {
		t.Fatalf("公開中の取得に失敗: %v", err)
	}
	if got.ID != published.ID {
		t.Errorf("id: %d を期待したが %d", published.ID, got.ID)
	}

	for name, id := range map[string]int64{
		"draft":  draft.ID,
		"closed": closed.ID,
		"不存在":    99999999,
	} {
		if _, err := uc.GetPublished(ctx, id); !errors.Is(err, usecase.ErrProjectNotFound) {
			t.Errorf("%s: ErrProjectNotFound を期待したが: %v", name, err)
		}
	}
}
