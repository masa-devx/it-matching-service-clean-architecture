package apitest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/e2efixture"
)

// projectPage は talent 一覧レスポンスの検証用の最小形
type projectPage struct {
	Projects []struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	} `json:"projects"`
	NextCursor *int64 `json:"next_cursor"`
}

// listProjects は talent の案件一覧をクエリ付きで取得する
func listProjects(t *testing.T, srv *httptest.Server, token, query string) projectPage {
	t.Helper()
	resp := Do(t, srv, http.MethodGet, "/talent/projects"+query, token, "")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("一覧: status 200 を期待したが %d", resp.StatusCode)
	}
	var page projectPage
	DecodeBody(t, resp, &page)
	return page
}

// ids はレスポンスの案件 ID 列を取り出す
func (p projectPage) ids() []int64 {
	out := make([]int64, 0, len(p.Projects))
	for _, pr := range p.Projects {
		out = append(out, pr.ID)
	}
	return out
}

// TestTalentProjectList は「公開中の案件だけが見える」の通しを固定する。
//
// 目的: 「未公開・原文は『取得しない』（画面で隠さない）」の原則が talent API で
// 効いていることを保証する。壊れると下書き・募集終了の案件が求職者に漏れる。
//
// 観点: 公開3件のみ（draft / closed が含まれない）・id 降順・所有企業をまたいで見える
// （B社の公開案件も出る＝company 一覧との対比）。
func TestTalentProjectList(t *testing.T) {
	srv, _ := NewServer(t)
	token := LoginTalentA(t, srv)

	page := listProjects(t, srv, token, "")

	want := []int64{e2efixture.Projects.BPublished, e2efixture.Projects.APublished2, e2efixture.Projects.APublished}
	got := page.ids()
	if len(got) != len(want) {
		t.Fatalf("公開3件を期待したが %d 件: %v", len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("projects[%d]: id %d を期待したが %d", i, want[i], got[i])
		}
	}
	for _, id := range got {
		if id == e2efixture.Projects.ADraft || id == e2efixture.Projects.AClosed {
			t.Errorf("未公開の案件 %d が talent に見えている", id)
		}
	}
	if page.NextCursor != nil {
		t.Errorf("3件は1ページに収まるので next_cursor は null のはずが %d", *page.NextCursor)
	}
}

// TestTalentProjectSeekPagination は seek ページネーションの通しを固定する。
//
// 目的: cursor の受け渡し（HTTP クエリ → handler → SQL）が通しで機能すること。
// 「次ページあり」と「最終ページ」を両側から張る（参考実装の NotHavingNextPage /
// HavingNextPage の型）。
//
// 観点: limit=2 で先頭2件 + next_cursor が3件目の id ／ cursor を渡すと続きが読めて
// 終端では next_cursor が null ／ limit=0 など不正値は既定にクランプされ全件返る。
func TestTalentProjectSeekPagination(t *testing.T) {
	srv, _ := NewServer(t)
	token := LoginTalentA(t, srv)

	// 1ページ目: id 降順の先頭2件（B社公開=5, A社公開2=3）+ 次ページの起点（=2）
	first := listProjects(t, srv, token, "?limit=2")
	if got := first.ids(); len(got) != 2 || got[0] != e2efixture.Projects.BPublished || got[1] != e2efixture.Projects.APublished2 {
		t.Fatalf("1ページ目: [5 3] を期待したが %v", got)
	}
	if first.NextCursor == nil {
		t.Fatal("次ページがあるのに next_cursor が null")
	}
	if *first.NextCursor != e2efixture.Projects.APublished {
		t.Fatalf("next_cursor: %d を期待したが %d", e2efixture.Projects.APublished, *first.NextCursor)
	}

	// 2ページ目: cursor から続きを読む → 残り1件・終端（next_cursor は null）
	second := listProjects(t, srv, token, fmt.Sprintf("?limit=2&cursor=%d", *first.NextCursor))
	if got := second.ids(); len(got) != 1 || got[0] != e2efixture.Projects.APublished {
		t.Fatalf("2ページ目: [2] を期待したが %v", got)
	}
	if second.NextCursor != nil {
		t.Errorf("最終ページなのに next_cursor が %d", *second.NextCursor)
	}

	// 不正な limit は既定値にクランプされ、エラーにならず全件返る
	clamped := listProjects(t, srv, token, "?limit=0")
	if len(clamped.Projects) != 3 {
		t.Errorf("limit=0 は既定値扱いで3件を期待したが %d 件", len(clamped.Projects))
	}
}

// TestTalentProjectSearch は検索フィルタの通しを固定する。
//
// 目的: クエリパラメータの解釈（カンマ区切り配列・bool・数値）が handler で正しく
// パースされ SQL まで届くこと。境界の網羅は usecase テストの仕事なので、ここでは
// 各フィルタの「ヒットする/しない」を対で張るだけに留める（層の責務）。
//
// 観点: skills は AND（全部持つ案件だけ）／remote_ok／min_hourly_rate。
// 検索対象のバリエーションは company API 経由で作成・公開する（アプリが作れるデータだけで
// 検索を検証する。1テスト限りのデータなので fixture には足さない）。
func TestTalentProjectSearch(t *testing.T) {
	srv, _ := NewServer(t)
	companyToken := LoginCompanyA(t, srv)
	talentToken := LoginTalentA(t, srv)

	// バリエーション案件を API 経由で作成 → 公開（React/TS・常駐・時給8000〜）
	body := `{"title":"React案件","description":"検索用","hours_per_week":10,"remote_ok":false,"required_skills":["React","TypeScript"],"hourly_rate_min":8000,"hourly_rate_max":9000}`
	resp := Do(t, srv, http.MethodPost, "/company/projects", companyToken, body)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("バリエーション案件の作成に失敗: %d", resp.StatusCode)
	}
	var created struct {
		ID int64 `json:"id"`
	}
	DecodeBody(t, resp, &created)
	resp = Do(t, srv, http.MethodPost, fmt.Sprintf("/company/projects/%d/publish", created.ID), companyToken, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("バリエーション案件の公開に失敗: %d", resp.StatusCode)
	}
	_ = resp.Body.Close()

	tests := []struct {
		name    string
		query   string
		wantIDs []int64
	}{
		{
			name:    "skills 単体: React を持つのは新規案件だけ",
			query:   "?skills=React",
			wantIDs: []int64{created.ID},
		},
		{
			name:    "skills は AND: Go と PostgreSQL 両方を持つ基準世界の3件のみ",
			query:   "?skills=Go,PostgreSQL",
			wantIDs: []int64{e2efixture.Projects.BPublished, e2efixture.Projects.APublished2, e2efixture.Projects.APublished},
		},
		{
			name:    "skills の AND でヒットなし: Go と React を両方持つ案件は無い",
			query:   "?skills=Go,React",
			wantIDs: []int64{},
		},
		{
			name:    "remote_ok=true: リモート可の基準世界3件のみ（常駐の新規案件は除外）",
			query:   "?remote_ok=true",
			wantIDs: []int64{e2efixture.Projects.BPublished, e2efixture.Projects.APublished2, e2efixture.Projects.APublished},
		},
		{
			name:    "min_hourly_rate=5000: 時給下限を設定済みの新規案件だけ（未設定は含まれない）",
			query:   "?min_hourly_rate=5000",
			wantIDs: []int64{created.ID},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := listProjects(t, srv, talentToken, tt.query)
			got := page.ids()
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("%d 件（%v）を期待したが %d 件（%v）", len(tt.wantIDs), tt.wantIDs, len(got), got)
			}
			for i := range tt.wantIDs {
				if got[i] != tt.wantIDs[i] {
					t.Errorf("[%d]: id %d を期待したが %d", i, tt.wantIDs[i], got[i])
				}
			}
		})
	}
}

// TestTalentProjectGet は案件詳細の公開チェックを固定する。
//
// 目的: 詳細エンドポイントでも「公開中だけ」が守られ、未公開と不存在が同じ 404
// （存在の有無を漏らさない）であることを保証する。
//
// 観点: 公開中 200（中身も確認）／draft 404／closed 404／不存在 404。
func TestTalentProjectGet(t *testing.T) {
	srv, _ := NewServer(t)
	token := LoginTalentA(t, srv)

	// 公開中: 200 + 中身
	resp := Do(t, srv, http.MethodGet, fmt.Sprintf("/talent/projects/%d", e2efixture.Projects.APublished), token, "")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("公開中: status 200 を期待したが %d", resp.StatusCode)
	}
	var detail struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}
	DecodeBody(t, resp, &detail)
	if detail.ID != e2efixture.Projects.APublished || detail.Title == "" {
		t.Errorf("詳細の中身が不正: %+v", detail)
	}

	// 未公開・不存在はすべて同じ 404
	tests := []struct {
		name      string
		projectID int64
	}{
		{"draft は 404（他社かどうか以前に未公開）", e2efixture.Projects.ADraft},
		{"closed は 404（募集終了は見せない）", e2efixture.Projects.AClosed},
		{"不存在も同じ 404", 99999999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := Do(t, srv, http.MethodGet, fmt.Sprintf("/talent/projects/%d", tt.projectID), token, "")
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusNotFound {
				t.Fatalf("status 404 を期待したが %d", resp.StatusCode)
			}
		})
	}
}
