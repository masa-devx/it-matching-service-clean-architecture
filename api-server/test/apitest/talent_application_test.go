package apitest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/e2efixture"
)

// applicationBody は応募レスポンスの検証用の最小形
type applicationBody struct {
	ID           int64  `json:"id"`
	ProjectID    int64  `json:"project_id"`
	ProjectTitle string `json:"project_title"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

// createPublishedProject は company API 経由で公開済みの案件を作る（検証用バリエーション）。
// 基準世界では両 talent が公開3件すべてに応募済み（UNIQUE 制約）のため、新規応募のテストには
// 新しい公開案件が必要になる
func createPublishedProject(t *testing.T, srv *httptest.Server, companyToken string) int64 {
	t.Helper()
	body := `{"title":"応募テスト用の新案件","description":"apitest","hours_per_week":10,"remote_ok":true,"required_skills":["Go"]}`
	resp := Do(t, srv, http.MethodPost, "/company/projects", companyToken, body)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("案件作成に失敗: %d", resp.StatusCode)
	}
	var created struct {
		ID int64 `json:"id"`
	}
	DecodeBody(t, resp, &created)

	pub := Do(t, srv, http.MethodPost, fmt.Sprintf("/company/projects/%d/publish", created.ID), companyToken, "")
	defer func() { _ = pub.Body.Close() }()
	if pub.StatusCode != http.StatusOK {
		t.Fatalf("案件公開に失敗: %d", pub.StatusCode)
	}
	return created.ID
}

// TestTalentApply は応募の作成の通しを固定する。
//
// 目的: 「公開中の案件にのみ応募できる」（INSERT...SELECT の原子的チェック）と
// 「二重応募は UNIQUE 制約で 409」が HTTP まで通しで効いていることを保証する。
//
// 観点: 公開案件へ 201 + DB 突合（applied で入る・talent_id はトークン由来）／
// 二重応募 409／draft への応募は 404（公開されていない案件の存在を漏らさない）。
func TestTalentApply(t *testing.T) {
	ctx := context.Background()
	srv, queries := NewServer(t)
	companyToken := LoginCompanyA(t, srv)
	talentToken := LoginTalentA(t, srv)

	// 新しい公開案件へ応募 → 201 + DB 突合
	projectID := createPublishedProject(t, srv, companyToken)
	resp := Do(t, srv, http.MethodPost, "/talent/applications", talentToken,
		fmt.Sprintf(`{"project_id":%d,"message":"ぜひやらせてください"}`, projectID))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("応募: status 201 を期待したが %d", resp.StatusCode)
	}
	var created applicationBody
	DecodeBody(t, resp, &created)
	if created.Status != "applied" || created.ProjectID != projectID {
		t.Errorf("レスポンスが不正: %+v", created)
	}
	row, err := queries.GetApplicationForTalent(ctx, db.GetApplicationForTalentParams{ID: created.ID, TalentID: e2efixture.TalentA.TalentID})
	if err != nil {
		t.Fatalf("作成した応募が DB から引けない（talent_id がトークン由来になっていない疑い）: %v", err)
	}
	if row.Status != "applied" || row.Message != "ぜひやらせてください" {
		t.Errorf("DB の実状態が不正: %+v", row)
	}

	// draft の案件への応募は 404（公開チェック。INSERT...SELECT の0行は SQL エラーではないので Tx を汚さない）
	resp = Do(t, srv, http.MethodPost, "/talent/applications", talentToken,
		fmt.Sprintf(`{"project_id":%d,"message":"未公開へ"}`, e2efixture.Projects.ADraft))
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("draft への応募: status 404 を期待したが %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 二重応募は 409。UNIQUE 違反は本物の SQL エラーで、テストを包む Tx を aborted にする
	// （25P02・本番は包む Tx が無いので無害）ため、このケースは必ずテストの最後に置く
	resp = Do(t, srv, http.MethodPost, "/talent/applications", talentToken,
		fmt.Sprintf(`{"project_id":%d,"message":"二回目"}`, projectID))
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("二重応募: status 409 を期待したが %d", resp.StatusCode)
	}
}

// TestTalentApplicationList は自分の応募一覧の通しを固定する。
//
// 目的: 「自分の応募だけが返る」（トークンの talent_id → WHERE）と、一覧表示に必要な
// project_title の JOIN 供給が通しで効いていることを保証する。
//
// 観点: TalentA の3件のみ・id 降順（TalentB の応募が混ざらない）・project_title が入る。
func TestTalentApplicationList(t *testing.T) {
	srv, _ := NewServer(t)
	token := LoginTalentA(t, srv)

	resp := Do(t, srv, http.MethodGet, "/talent/applications", token, "")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status 200 を期待したが %d", resp.StatusCode)
	}
	var body struct {
		Applications []applicationBody `json:"applications"`
	}
	DecodeBody(t, resp, &body)

	// TalentA の応募 = Applied(1)・Accepted(3)・Withdrawn(5) → id 降順
	want := []int64{e2efixture.Applications.Withdrawn, e2efixture.Applications.Accepted, e2efixture.Applications.Applied}
	if len(body.Applications) != len(want) {
		t.Fatalf("3件を期待したが %d 件", len(body.Applications))
	}
	for i, app := range body.Applications {
		if app.ID != want[i] {
			t.Errorf("applications[%d].id: %d を期待したが %d", i, want[i], app.ID)
		}
		if app.ProjectTitle == "" {
			t.Errorf("applications[%d] に project_title が無い", i)
		}
	}
}

// TestTalentTransitions は talent 側の遷移（withdraw / accept / decline）の通しを固定する。
//
// 目的: 遷移表（actor=talent）が HTTP まで通しで効いていること
// （*_acted_at の記録は usecase テストで担保済み・ここでは繰り返さない）。
// accept と decline は「同じ基準世界の Offered(id=2) を別々のテスト実行が別々の運命に導く」
// ＝ Tx 分離により各テストが独立した世界を持つことの実演でもある。
//
// 観点: applied→withdrawn／offered→accepted（ダブルオプトイン成立）／offered→declined。
// レスポンスと DB の両方を突合する。
func TestTalentTransitions(t *testing.T) {
	tests := []struct {
		name          string
		loginEmail    string
		applicationID int64
		action        string
		wantStatus    string
		talentID      int64
	}{
		{"withdraw: applied → withdrawn", e2efixture.TalentA.Email, e2efixture.Applications.Applied, "withdraw", "withdrawn", e2efixture.TalentA.TalentID},
		{"accept: offered → accepted（ダブルオプトイン成立）", e2efixture.TalentB.Email, e2efixture.Applications.Offered, "accept", "accepted", e2efixture.TalentB.TalentID},
		{"decline: 同じ offered を別世界で辞退に導く", e2efixture.TalentB.Email, e2efixture.Applications.Offered, "decline", "declined", e2efixture.TalentB.TalentID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			srv, queries := NewServer(t) // テストごとに独立した世界（同じ id=2 を使い回せる理由）
			token := Login(t, srv, "/talent", tt.loginEmail)

			resp := Do(t, srv, http.MethodPost, fmt.Sprintf("/talent/applications/%d/%s", tt.applicationID, tt.action), token, "")
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status 200 を期待したが %d", resp.StatusCode)
			}
			var body applicationBody
			DecodeBody(t, resp, &body)
			if body.Status != tt.wantStatus {
				t.Errorf("レスポンスの status: %s を期待したが %s", tt.wantStatus, body.Status)
			}

			row, err := queries.GetApplicationForTalent(ctx, db.GetApplicationForTalentParams{ID: tt.applicationID, TalentID: tt.talentID})
			if err != nil {
				t.Fatalf("応募が DB から引けない: %v", err)
			}
			if row.Status != tt.wantStatus {
				t.Errorf("DB の status: %s を期待したが %s", tt.wantStatus, row.Status)
			}
		})
	}
}

// TestTalentTransitionGuards は talent 側の遷移の異常系を固定する。
//
// 目的: 終端状態の保護（決着の巻き戻し禁止）と所有チェックが HTTP まで効いていること。
// 409 のエラーメッセージに「現在の状態」が含まれる契約（画面がエラー文言をそのまま出せる）も固定する。
//
// 観点: accepted への withdraw は 409 + メッセージに accepted／他人の応募への accept は
// 404（存在の有無を漏らさない）／不存在も同じ 404。
func TestTalentTransitionGuards(t *testing.T) {
	srv, _ := NewServer(t)
	tokenA := LoginTalentA(t, srv)

	// 終端（accepted）への操作は 409
	resp := Do(t, srv, http.MethodPost, fmt.Sprintf("/talent/applications/%d/withdraw", e2efixture.Applications.Accepted), tokenA, "")
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("終端への操作: status 409 を期待したが %d", resp.StatusCode)
	}
	var conflict struct {
		Error string `json:"error"`
	}
	DecodeBody(t, resp, &conflict)
	if !strings.Contains(conflict.Error, "accepted") {
		t.Errorf("エラーメッセージに現在の状態が含まれない: %q", conflict.Error)
	}

	// 他人（TalentB）の応募への操作は 404
	resp = Do(t, srv, http.MethodPost, fmt.Sprintf("/talent/applications/%d/accept", e2efixture.Applications.Offered), tokenA, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("他人の応募: status 404 を期待したが %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 不存在も同じ 404
	resp = Do(t, srv, http.MethodPost, "/talent/applications/99999999/withdraw", tokenA, "")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("不存在: status 404 を期待したが %d", resp.StatusCode)
	}
}
