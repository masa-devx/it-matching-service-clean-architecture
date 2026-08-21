package apitest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/e2efixture"
)

// TestCompanyProjectList は company の案件一覧の通しを固定する。
//
// 目的: 「自社の案件だけが返る」（テナント分離）が、JWT の company_id → SQL の WHERE まで
// 通しで効いていることを保証する。壊れると他社の下書き・選考状況が漏れる。
//
// 観点: 自社4件のみ（B社の公開案件が混ざらない）・id 降順。
// 件数と ID 集合の両方で「多すぎない・欠けない」を固定する。
func TestCompanyProjectList(t *testing.T) {
	srv, _ := NewServer(t)
	token := LoginCompanyA(t, srv)

	resp := Do(t, srv, http.MethodGet, "/company/projects", token, "")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status 200 を期待したが %d", resp.StatusCode)
	}
	var body struct {
		Projects []struct {
			ID     int64  `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"projects"`
	}
	DecodeBody(t, resp, &body)

	if len(body.Projects) != 4 {
		t.Fatalf("自社4件を期待したが %d 件", len(body.Projects))
	}
	// id 降順（4,3,2,1）= B社の案件(5)が先頭に混ざればここで検出される
	wantIDs := []int64{e2efixture.Projects.AClosed, e2efixture.Projects.APublished2, e2efixture.Projects.APublished, e2efixture.Projects.ADraft}
	for i, p := range body.Projects {
		if p.ID != wantIDs[i] {
			t.Errorf("projects[%d].id: %d を期待したが %d", i, wantIDs[i], p.ID)
		}
		if p.ID == e2efixture.Projects.BPublished {
			t.Errorf("他社（B社）の案件 %d が一覧に混ざっている", p.ID)
		}
	}
}

// TestCompanyProjectGet は案件詳細の所有チェックを固定する。
//
// 目的: 「所有チェックは SQL の WHERE に埋め込む」が HTTP まで通しで効いていること。
// 他社と不存在が同じ 404（存在の有無を漏らさない）であることも固定する。
//
// 観点: 自社 draft は 200（未公開でも自社なら見える）／他社の公開案件 404／不存在 404。
func TestCompanyProjectGet(t *testing.T) {
	srv, _ := NewServer(t)
	token := LoginCompanyA(t, srv)

	tests := []struct {
		name       string
		projectID  int64
		wantStatus int
	}{
		{"自社の draft は 200（未公開でも自社なら見える）", e2efixture.Projects.ADraft, http.StatusOK},
		{"他社の公開案件は 404（公開されていても company API では他社のもの）", e2efixture.Projects.BPublished, http.StatusNotFound},
		{"不存在も同じ 404（存在の有無を漏らさない）", 99999999, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := Do(t, srv, http.MethodGet, fmt.Sprintf("/company/projects/%d", tt.projectID), token, "")
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status %d を期待したが %d", tt.wantStatus, resp.StatusCode)
			}
		})
	}
}

// TestCompanyProjectCreateUpdate は作成・更新の通しを固定する。
//
// 目的: 生成型 → DB の詰め替え（handler の固有責務）が正しいこと。
// レスポンスと DB の実状態を両方突合する（観点ルール）。
// 作成は「基準世界の固定 ID とシーケンスが衝突しない」ことの実地確認も兼ねる。
//
// 観点: 作成 201 + DB に draft で入る／更新 200 + DB 反映／他社の更新は 404。
func TestCompanyProjectCreateUpdate(t *testing.T) {
	ctx := context.Background()
	srv, queries := NewServer(t)
	token := LoginCompanyA(t, srv)

	// 作成: 201 + DB 突合
	createBody := `{"title":"新規の案件","description":"統合テストから作成","hours_per_week":15,"remote_ok":true,"required_skills":["Go"]}`
	resp := Do(t, srv, http.MethodPost, "/company/projects", token, createBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("作成: status 201 を期待したが %d", resp.StatusCode)
	}
	var created struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	DecodeBody(t, resp, &created)
	if created.Status != "draft" {
		t.Errorf("作成直後は draft を期待したが %s", created.Status)
	}
	if created.ID <= e2efixture.Projects.BPublished {
		t.Errorf("新規 ID %d が基準世界の範囲（〜%d）と重なっている", created.ID, e2efixture.Projects.BPublished)
	}
	row, err := queries.GetProjectForCompany(ctx, db.GetProjectForCompanyParams{ID: created.ID, CompanyID: e2efixture.CompanyA.CompanyID})
	if err != nil {
		t.Fatalf("作成した案件が DB から引けない: %v", err)
	}
	if row.Title != "新規の案件" || row.Status != "draft" {
		t.Errorf("DB の実状態が入力と一致しない: %+v", row)
	}

	// 更新（仕様は PATCH）: 200 + DB 反映
	updateBody := `{"title":"更新後のタイトル","description":"更新済み","hours_per_week":20,"remote_ok":false,"required_skills":["Go","SQL"]}`
	resp = Do(t, srv, http.MethodPatch, fmt.Sprintf("/company/projects/%d", created.ID), token, updateBody)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("更新: status 200 を期待したが %d", resp.StatusCode)
	}
	_ = resp.Body.Close()
	row, err = queries.GetProjectForCompany(ctx, db.GetProjectForCompanyParams{ID: created.ID, CompanyID: e2efixture.CompanyA.CompanyID})
	if err != nil {
		t.Fatalf("更新後の案件が DB から引けない: %v", err)
	}
	if row.Title != "更新後のタイトル" || row.HoursPerWeek != 20 {
		t.Errorf("更新が DB に反映されていない: %+v", row)
	}

	// 他社の案件の更新は 404
	resp = Do(t, srv, http.MethodPatch, fmt.Sprintf("/company/projects/%d", e2efixture.Projects.BPublished), token, updateBody)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("他社更新: status 404 を期待したが %d", resp.StatusCode)
	}
}

// TestCompanyProjectTransition は公開・クローズ遷移の通しを固定する。
//
// 目的: 遷移表（shared/domain）が HTTP 経由でも守られていること。
// 正常遷移はレスポンスと DB の両方で状態を確認し、不正遷移は 409 で止まることを固定する。
// 「closed → published は再募集として許可」という遷移表の意図も、ここで実行可能な文書にする。
//
// 観点: draft→close の飛び級は 409（公開を経ずに閉じられない）／draft→published→closed の
// 正常系／closed→published の再公開は 200（再募集）／他社の案件の publish は 404。
func TestCompanyProjectTransition(t *testing.T) {
	ctx := context.Background()
	srv, queries := NewServer(t)
	token := LoginCompanyA(t, srv)

	// draft → closed の飛び級は 409（現在の状態がエラーメッセージで分かる契約）
	id := e2efixture.Projects.ADraft
	resp := Do(t, srv, http.MethodPost, fmt.Sprintf("/company/projects/%d/close", id), token, "")
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("draft→close: status 409 を期待したが %d", resp.StatusCode)
	}
	var conflict struct {
		Error string `json:"error"`
	}
	DecodeBody(t, resp, &conflict)
	if conflict.Error == "" {
		t.Error("409 のエラーメッセージが空")
	}

	// draft → published → closed（基準世界の draft を使って通しで進める）
	resp = Do(t, srv, http.MethodPost, fmt.Sprintf("/company/projects/%d/publish", id), token, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("publish: status 200 を期待したが %d", resp.StatusCode)
	}
	_ = resp.Body.Close()
	resp = Do(t, srv, http.MethodPost, fmt.Sprintf("/company/projects/%d/close", id), token, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("close: status 200 を期待したが %d", resp.StatusCode)
	}
	_ = resp.Body.Close()
	row, err := queries.GetProjectForCompany(ctx, db.GetProjectForCompanyParams{ID: id, CompanyID: e2efixture.CompanyA.CompanyID})
	if err != nil {
		t.Fatalf("遷移後の案件が DB から引けない: %v", err)
	}
	if row.Status != "closed" {
		t.Errorf("DB の status: closed を期待したが %s", row.Status)
	}

	// closed → published は再募集として 200（遷移表の意図を文書化）
	resp = Do(t, srv, http.MethodPost, fmt.Sprintf("/company/projects/%d/publish", e2efixture.Projects.AClosed), token, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("再募集（closed→published）: status 200 を期待したが %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 他社の案件の publish は 404
	resp = Do(t, srv, http.MethodPost, fmt.Sprintf("/company/projects/%d/publish", e2efixture.Projects.BPublished), token, "")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("他社publish: status 404 を期待したが %d", resp.StatusCode)
	}
}

// TestCompanyProjectApplications は自社案件の応募一覧の通しを固定する。
//
// 目的: JOIN による応募者プロフィールの供給（talent_display_name / talent_skills）が
// HTTP まで届くこと・他社案件の応募一覧が 404 で守られていることを保証する。
//
// 観点: 公開案件1の応募2件（applied + offered）が新しい順・プロフィール込みで返る／
// 他社（B社）の案件の応募一覧は 404。
func TestCompanyProjectApplications(t *testing.T) {
	srv, _ := NewServer(t)
	token := LoginCompanyA(t, srv)

	resp := Do(t, srv, http.MethodGet, fmt.Sprintf("/company/projects/%d/applications", e2efixture.Projects.APublished), token, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status 200 を期待したが %d", resp.StatusCode)
	}
	var body struct {
		Applications []struct {
			ID                int64    `json:"id"`
			Status            string   `json:"status"`
			TalentDisplayName string   `json:"talent_display_name"`
			TalentSkills      []string `json:"talent_skills"`
		} `json:"applications"`
	}
	DecodeBody(t, resp, &body)

	if len(body.Applications) != 2 {
		t.Fatalf("2件を期待したが %d 件", len(body.Applications))
	}
	// id 降順: offered(2) → applied(1)
	if body.Applications[0].ID != e2efixture.Applications.Offered || body.Applications[1].ID != e2efixture.Applications.Applied {
		t.Errorf("新しい順を期待したが: %d, %d", body.Applications[0].ID, body.Applications[1].ID)
	}
	if body.Applications[0].TalentDisplayName == "" || len(body.Applications[0].TalentSkills) == 0 {
		t.Errorf("応募者プロフィールが載っていない: %+v", body.Applications[0])
	}

	// 他社の案件の応募一覧は 404
	resp = Do(t, srv, http.MethodGet, fmt.Sprintf("/company/projects/%d/applications", e2efixture.Projects.BPublished), token, "")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("他社の応募一覧: status 404 を期待したが %d", resp.StatusCode)
	}
}
