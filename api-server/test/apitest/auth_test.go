package apitest

import (
	"net/http"
	"testing"
)

// TestAuthz は認可の通し（本番と同一配線・実ミドルウェア・本物の JWT）を固定する。
//
// 目的: 「ロール認可はパスプレフィックス × ミドルウェアで一律」という設計の見せ場そのものを
// 守る。壊れると未認証アクセスやロール越境（talent が企業の応募一覧を見る等）が通ってしまう。
// このリポジトリで唯一、認可ミドルウェアを通したリクエストを検証する場所。
//
// 観点: 401（誰か分からない）と 403（誰かは分かるが権限が無い）の区別を両ロール対称に張る。
// 401 はヘッダー無し・改ざんトークンの両方（失敗理由が同一レスポンスに潰されること＝情報を
// 漏らさない、も含めて固定）。認証不要の運用エンドポイント（/health）が素通りすることも固定する。
func TestAuthz(t *testing.T) {
	srv, _ := NewServer(t)
	companyToken := LoginCompanyA(t, srv)
	talentToken := LoginTalentA(t, srv)

	tests := []struct {
		name       string
		method     string
		path       string
		token      string
		wantStatus int
		wantError  string // 空なら本文は見ない
	}{
		{
			name:   "未認証で company の保護APIは 401",
			method: http.MethodGet, path: "/company/projects",
			token: "", wantStatus: http.StatusUnauthorized, wantError: "認証が必要です",
		},
		{
			name:   "未認証で talent の保護APIは 401",
			method: http.MethodGet, path: "/talent/projects",
			token: "", wantStatus: http.StatusUnauthorized, wantError: "認証が必要です",
		},
		{
			name:   "改ざんトークンは 401（ヘッダー無しと同じ応答＝理由を漏らさない）",
			method: http.MethodGet, path: "/company/projects",
			token: "tampered.token.value", wantStatus: http.StatusUnauthorized, wantError: "認証が必要です",
		},
		{
			name:   "talent トークンで company API は 403",
			method: http.MethodGet, path: "/company/projects",
			token: "talent", wantStatus: http.StatusForbidden, wantError: "この操作を行う権限がありません",
		},
		{
			name:   "company トークンで talent API は 403",
			method: http.MethodGet, path: "/talent/projects",
			token: "company", wantStatus: http.StatusForbidden, wantError: "この操作を行う権限がありません",
		},
		{
			name:   "company の正規トークンは 200",
			method: http.MethodGet, path: "/company/projects",
			token: "company", wantStatus: http.StatusOK,
		},
		{
			name:   "talent の正規トークンは 200",
			method: http.MethodGet, path: "/talent/projects",
			token: "talent", wantStatus: http.StatusOK,
		},
		{
			name:   "/health は認証不要（運用エンドポイント）",
			method: http.MethodGet, path: "/health",
			token: "", wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// トークン欄の "company" / "talent" は本物のログイン済みトークンに展開する
			token := tt.token
			switch token {
			case "company":
				token = companyToken
			case "talent":
				token = talentToken
			}

			resp := Do(t, srv, tt.method, tt.path, token, "")
			if resp.StatusCode != tt.wantStatus {
				_ = resp.Body.Close()
				t.Fatalf("status %d を期待したが %d", tt.wantStatus, resp.StatusCode)
			}

			if tt.wantError == "" {
				_ = resp.Body.Close()
				return
			}
			var body struct {
				Error string `json:"error"`
			}
			DecodeBody(t, resp, &body)
			if body.Error != tt.wantError {
				t.Errorf("エラーメッセージ %q を期待したが %q", tt.wantError, body.Error)
			}
		})
	}
}

// TestAuthzParallel は統合テスト基盤が並列実行に耐えることを固定する。
//
// 目的: 「サーバーごとに独立した Tx + 基準世界」という基盤設計の約束
// （後始末不要・並列可）が、httptest まで含めても成り立つことを保証する。
// 観点: 2つのサーバーを並列に立て、それぞれでログイン〜保護APIの取得が完結する。
func TestAuthzParallel(t *testing.T) {
	for i := range 2 {
		t.Run([]string{"server-0", "server-1"}[i], func(t *testing.T) {
			t.Parallel()
			srv, _ := NewServer(t)
			token := LoginCompanyA(t, srv)

			resp := Do(t, srv, http.MethodGet, "/company/projects", token, "")
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status 200 を期待したが %d", resp.StatusCode)
			}
		})
	}
}
