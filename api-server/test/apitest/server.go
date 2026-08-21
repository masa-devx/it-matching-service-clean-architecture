// Package apitest は API 統合テストの共通基盤を置く（#93）。
//
// 方針: 本番と同一の配線（app.NewRouter）を httptest で起動し、認証は本物の
// signin API から得た JWT を使う（ミドルウェアやトークンをモックしない）。
// データは基準世界（e2efixture）+ テスト固有分は factories。検証は
// 「レスポンスと DB の実状態の両方」を突合する（.claude/rules/backend.md の観点ルール）
package apitest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/app"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/e2efixture"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/helpers"
)

// jwtSecret はテスト専用の署名鍵（本番とは無関係の公開値）
const jwtSecret = "apitest-only-secret"

// NewServer は「基準世界入りのトランザクション」の上に本番と同一配線のサーバーを立てる。
// サーバーもトランザクションもテスト終了時に自動で片付く（Close / ROLLBACK）
func NewServer(t *testing.T) (*httptest.Server, *db.Queries) {
	t.Helper()

	tx, queries := helpers.NewTestTx(t)
	e2efixture.Load(t, tx)

	handler, err := app.NewRouter(tx, queries, []byte(jwtSecret))
	if err != nil {
		t.Fatalf("ルーターの組み立てに失敗: %v", err)
	}
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv, queries
}

// Login は本物のログイン API を叩いて Bearer トークンを得る。
// rolePath は "/company" か "/talent"。認証の経路そのものがテストの前提として毎回検証される
func Login(t *testing.T, srv *httptest.Server, rolePath, email string) string {
	t.Helper()

	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, e2efixture.Password)
	resp := Do(t, srv, http.MethodPost, rolePath+"/auth/login", "", body)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("%s のログインに失敗: status %d", email, resp.StatusCode)
	}

	var parsed struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("ログインレスポンスの解析に失敗: %v", err)
	}
	if parsed.Token == "" {
		t.Fatal("ログインレスポンスに token が無い")
	}
	return parsed.Token
}

// LoginCompanyA / LoginTalentA は基準世界の代表ユーザーでログインする
func LoginCompanyA(t *testing.T, srv *httptest.Server) string {
	return Login(t, srv, "/company", e2efixture.CompanyA.Email)
}

func LoginTalentA(t *testing.T, srv *httptest.Server) string {
	return Login(t, srv, "/talent", e2efixture.TalentA.Email)
}

// Do はトークン付きでリクエストを送る（token が空なら未認証リクエスト）。
// Body の Close は呼び出し側が行う
func Do(t *testing.T, srv *httptest.Server, method, path, token, body string) *http.Response {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, srv.URL+path, reader)
	if err != nil {
		t.Fatalf("リクエストの作成に失敗: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s のリクエストに失敗: %v", method, path, err)
	}
	return resp
}

// DecodeBody はレスポンスボディを JSON として読み、Body を閉じる
func DecodeBody(t *testing.T, resp *http.Response, into any) {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
		t.Fatalf("レスポンスボディの解析に失敗: %v", err)
	}
}
