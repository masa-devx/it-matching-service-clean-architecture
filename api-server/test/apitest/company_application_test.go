package apitest

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/e2efixture"
)

// TestCompanyOfferReject は company 側の選考遷移の通しを固定する。
//
// 目的: 遷移表（actor=company）と JOIN 越しの所有チェック（applications は company_id を
// 持たない）が HTTP まで通しで効いていることを保証する。
//
// 観点: offer（applied→offered）と reject（applied→rejected）— 同じ基準世界の
// Applied(id=1) を別々のテスト実行が別々の運命に導く。レスポンスに応募者プロフィールが
// 載ること・DB の実状態も突合する。
func TestCompanyOfferReject(t *testing.T) {
	tests := []struct {
		name       string
		action     string
		wantStatus string
	}{
		{"offer: applied → offered", "offer", "offered"},
		{"reject: 同じ applied を別世界で不採用に導く", "reject", "rejected"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			srv, queries := NewServer(t)
			token := LoginCompanyA(t, srv)

			resp := Do(t, srv, http.MethodPost, fmt.Sprintf("/company/applications/%d/%s", e2efixture.Applications.Applied, tt.action), token, "")
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status 200 を期待したが %d", resp.StatusCode)
			}
			var body struct {
				ID                int64  `json:"id"`
				Status            string `json:"status"`
				TalentDisplayName string `json:"talent_display_name"`
			}
			DecodeBody(t, resp, &body)
			if body.Status != tt.wantStatus {
				t.Errorf("レスポンスの status: %s を期待したが %s", tt.wantStatus, body.Status)
			}
			if body.TalentDisplayName == "" {
				t.Error("応募者プロフィール（talent_display_name）が載っていない")
			}

			row, err := queries.GetApplicationForCompany(ctx, db.GetApplicationForCompanyParams{ID: e2efixture.Applications.Applied, CompanyID: e2efixture.CompanyA.CompanyID})
			if err != nil {
				t.Fatalf("応募が DB から引けない: %v", err)
			}
			if row.Status != tt.wantStatus {
				t.Errorf("DB の status: %s を期待したが %s", tt.wantStatus, row.Status)
			}
		})
	}
}

// TestCompanyOfferGuards は company 側の選考の異常系を固定する。
//
// 目的: 二重オファーの 409 契約（エラーメッセージに現在の状態が含まれる）と、
// JOIN 越しの所有チェック（他社の案件への応募・不存在が同じ 404）を保証する。
//
// 観点: offered への再 offer は 409 + "offered"／他社（CompanyB のトークンで A 社案件の応募）
// は 404／不存在も同じ 404。いずれも SQL エラーではない（UPDATE...WHERE の0行）ため
// Tx を汚さず、順序の制約はない。
func TestCompanyOfferGuards(t *testing.T) {
	srv, _ := NewServer(t)
	tokenA := LoginCompanyA(t, srv)
	tokenB := Login(t, srv, "/company", e2efixture.CompanyB.Email)

	// 二重オファー: offered への再 offer は 409 + 現在の状態がメッセージに含まれる
	resp := Do(t, srv, http.MethodPost, fmt.Sprintf("/company/applications/%d/offer", e2efixture.Applications.Offered), tokenA, "")
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("二重オファー: status 409 を期待したが %d", resp.StatusCode)
	}
	var conflict struct {
		Error string `json:"error"`
	}
	DecodeBody(t, resp, &conflict)
	if !strings.Contains(conflict.Error, "offered") {
		t.Errorf("エラーメッセージに現在の状態が含まれない: %q", conflict.Error)
	}

	// 他社: CompanyB のトークンで A 社案件の応募に offer → 404（存在の有無を漏らさない）
	resp = Do(t, srv, http.MethodPost, fmt.Sprintf("/company/applications/%d/offer", e2efixture.Applications.Applied), tokenB, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("他社の応募: status 404 を期待したが %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 不存在も同じ 404
	resp = Do(t, srv, http.MethodPost, "/company/applications/99999999/offer", tokenA, "")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("不存在: status 404 を期待したが %d", resp.StatusCode)
	}
}

// TestDoubleOptInFlow はダブルオプトイン（応募 → 選考 → 承諾）のフル通しを固定する。
//
// 目的: Epic #90 の受け入れシナリオそのもの。案件の作成から承諾まで、2つのロールの
// 本物のトークンを行き来しながら、1つの応募の一生が API だけで完結することを保証する。
// 壊れるとこのサービスの中核フロー（マッチング成立）が成立しない。
//
// 観点: 各ステップでレスポンスの状態を確認し、最後に DB の実状態を**両視点の取得クエリ**で
// 突合する（acted_at の記録自体は usecase テストで担保済み・ここでは繰り返さない）。
func TestDoubleOptInFlow(t *testing.T) {
	ctx := context.Background()
	srv, queries := NewServer(t)
	companyToken := LoginCompanyA(t, srv)
	talentToken := LoginTalentA(t, srv)

	// 1. company: 案件を作成して公開
	projectID := createPublishedProject(t, srv, companyToken)

	// 2. talent: 応募する
	resp := Do(t, srv, http.MethodPost, "/talent/applications", talentToken,
		fmt.Sprintf(`{"project_id":%d,"message":"ダブルオプトインの通しです"}`, projectID))
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("応募: status 201 を期待したが %d", resp.StatusCode)
	}
	var application applicationBody
	DecodeBody(t, resp, &application)

	// 3. company: 応募一覧に現れたことを確認する
	resp = Do(t, srv, http.MethodGet, fmt.Sprintf("/company/projects/%d/applications", projectID), companyToken, "")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("応募一覧: status 200 を期待したが %d", resp.StatusCode)
	}
	var list struct {
		Applications []struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"applications"`
	}
	DecodeBody(t, resp, &list)
	if len(list.Applications) != 1 || list.Applications[0].ID != application.ID {
		t.Fatalf("応募一覧に今の応募が現れない: %+v", list.Applications)
	}

	// 4. company: オファーする
	resp = Do(t, srv, http.MethodPost, fmt.Sprintf("/company/applications/%d/offer", application.ID), companyToken, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("オファー: status 200 を期待したが %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	// 5. talent: 自分の一覧で offered になったことを確認する
	resp = Do(t, srv, http.MethodGet, "/talent/applications", talentToken, "")
	defer func() { _ = resp.Body.Close() }()
	var mine struct {
		Applications []applicationBody `json:"applications"`
	}
	DecodeBody(t, resp, &mine)
	found := false
	for _, app := range mine.Applications {
		if app.ID == application.ID {
			found = true
			if app.Status != "offered" {
				t.Errorf("talent 側の表示: offered を期待したが %s", app.Status)
			}
		}
	}
	if !found {
		t.Fatal("talent の一覧に応募が現れない")
	}

	// 6. talent: 承諾する（ダブルオプトイン成立）
	resp = Do(t, srv, http.MethodPost, fmt.Sprintf("/talent/applications/%d/accept", application.ID), talentToken, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("承諾: status 200 を期待したが %d", resp.StatusCode)
	}
	var accepted applicationBody
	DecodeBody(t, resp, &accepted)
	if accepted.Status != "accepted" {
		t.Errorf("承諾後の status: accepted を期待したが %s", accepted.Status)
	}

	// 7. DB の最終状態を両視点の取得クエリで突合する（acted_at の記録自体は usecase テストで担保済み）
	companyView, err := queries.GetApplicationForCompany(ctx, db.GetApplicationForCompanyParams{ID: application.ID, CompanyID: e2efixture.CompanyA.CompanyID})
	if err != nil {
		t.Fatalf("最終状態が company 視点で引けない: %v", err)
	}
	if companyView.Status != "accepted" {
		t.Errorf("company 視点の最終状態: accepted を期待したが %s", companyView.Status)
	}
	talentView, err := queries.GetApplicationForTalent(ctx, db.GetApplicationForTalentParams{ID: application.ID, TalentID: e2efixture.TalentA.TalentID})
	if err != nil {
		t.Fatalf("最終状態が talent 視点で引けない: %v", err)
	}
	if talentView.Status != "accepted" {
		t.Errorf("talent 視点の最終状態: accepted を期待したが %s", talentView.Status)
	}
}
