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

// TestOfferWithMessage はオファーメッセージの通し（company が添えて talent に届くまで）を固定する。
//
// 目的: #98 の機能の中核——「オファー時にのみメッセージを設定でき、応募者に届く」——を
// HTTP からレスポンス・DB・相手側の画面まで通しで保証する。
//
// 観点: メッセージ付き offer → レスポンスと DB に offer_message／talent の一覧に届く。
// 基準世界の Offered（メッセージあり）と Rejected（NULL）も fixture 経由の供給として突合する。
func TestOfferWithMessage(t *testing.T) {
	ctx := context.Background()
	srv, queries := NewServer(t)
	companyToken := LoginCompanyA(t, srv)

	// メッセージ付きオファー（基準世界の Applied を使用）
	const message = "経歴を拝見しました。ぜひ当社の案件でお願いしたいです。"
	resp := Do(t, srv, http.MethodPost, fmt.Sprintf("/company/applications/%d/offer", e2efixture.Applications.Applied), companyToken,
		fmt.Sprintf(`{"message":%q}`, message))
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status 200 を期待したが %d", resp.StatusCode)
	}
	var offered struct {
		Status       string  `json:"status"`
		OfferMessage *string `json:"offer_message"`
	}
	DecodeBody(t, resp, &offered)
	if offered.OfferMessage == nil || *offered.OfferMessage != message {
		t.Errorf("レスポンスの offer_message が入力と一致しない: %v", offered.OfferMessage)
	}

	row, err := queries.GetApplicationForCompany(ctx, db.GetApplicationForCompanyParams{ID: e2efixture.Applications.Applied, CompanyID: e2efixture.CompanyA.CompanyID})
	if err != nil {
		t.Fatalf("DB から引けない: %v", err)
	}
	if row.OfferMessage == nil || *row.OfferMessage != message {
		t.Errorf("DB の offer_message が入力と一致しない: %v", row.OfferMessage)
	}

	// talent（応募者 = TalentA）の一覧に届くこと
	talentToken := LoginTalentA(t, srv)
	listResp := Do(t, srv, http.MethodGet, "/talent/applications", talentToken, "")
	defer func() { _ = listResp.Body.Close() }()
	var mine struct {
		Applications []applicationBody `json:"applications"`
	}
	DecodeBody(t, listResp, &mine)
	for _, app := range mine.Applications {
		if app.ID == e2efixture.Applications.Applied {
			if app.OfferMessage == nil || *app.OfferMessage != message {
				t.Errorf("talent の一覧に offer_message が届いていない: %v", app.OfferMessage)
			}
		}
	}
}

// TestOfferMessageFixture は基準世界のオファーメッセージ供給を固定する。
//
// 目的: seed（一次情報）→ dump → Load の経路でメッセージが正しく世界に載っていること。
// NULL（メッセージ無し）と値の両方が基準世界に揃っている設計もここで守る。
//
// 観点: TalentB の一覧で Offered=定数どおり／Declined=B社の定数どおり／Rejected=NULL。
func TestOfferMessageFixture(t *testing.T) {
	srv, _ := NewServer(t)
	token := Login(t, srv, "/talent", e2efixture.TalentB.Email)

	resp := Do(t, srv, http.MethodGet, "/talent/applications", token, "")
	defer func() { _ = resp.Body.Close() }()
	var mine struct {
		Applications []applicationBody `json:"applications"`
	}
	DecodeBody(t, resp, &mine)

	offeredMsg, declinedMsg := e2efixture.OfferedMessage, e2efixture.DeclinedMessage
	want := map[int64]*string{
		e2efixture.Applications.Offered:  &offeredMsg,
		e2efixture.Applications.Declined: &declinedMsg,
		e2efixture.Applications.Rejected: nil,
	}
	for _, app := range mine.Applications {
		expected, ok := want[app.ID]
		if !ok {
			continue
		}
		switch {
		case expected == nil && app.OfferMessage != nil:
			t.Errorf("応募 %d: offer_message は NULL のはずが %q", app.ID, *app.OfferMessage)
		case expected != nil && (app.OfferMessage == nil || *app.OfferMessage != *expected):
			t.Errorf("応募 %d: offer_message %q を期待したが %v", app.ID, *expected, app.OfferMessage)
		}
	}
}

// TestOfferMessageValidation はメッセージ長の検証の通しを固定する。
//
// 目的: validator（501文字で拒否）が HTTP 経由でも効き、拒否時に遷移が起きないこと
// （バリデーションは遷移の前）を保証する。
//
// 観点: 501文字 → 400 + 利用者向けメッセージ／応募は applied のまま（DB 突合）。
func TestOfferMessageValidation(t *testing.T) {
	ctx := context.Background()
	srv, queries := NewServer(t)
	token := LoginCompanyA(t, srv)

	tooLong := strings.Repeat("あ", 501)
	resp := Do(t, srv, http.MethodPost, fmt.Sprintf("/company/applications/%d/offer", e2efixture.Applications.Applied), token,
		fmt.Sprintf(`{"message":%q}`, tooLong))
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status 400 を期待したが %d", resp.StatusCode)
	}
	var body struct {
		Error string `json:"error"`
	}
	DecodeBody(t, resp, &body)
	if !strings.Contains(body.Error, "500文字") {
		t.Errorf("利用者向けのエラーメッセージを期待したが: %q", body.Error)
	}

	// バリデーションで止まったので遷移は起きていない（applied のまま・offer_message も NULL）
	row, err := queries.GetApplicationForCompany(ctx, db.GetApplicationForCompanyParams{ID: e2efixture.Applications.Applied, CompanyID: e2efixture.CompanyA.CompanyID})
	if err != nil {
		t.Fatalf("DB から引けない: %v", err)
	}
	if row.Status != "applied" || row.OfferMessage != nil {
		t.Errorf("拒否後も状態が変わらないはず: status=%s offer_message=%v", row.Status, row.OfferMessage)
	}
}
