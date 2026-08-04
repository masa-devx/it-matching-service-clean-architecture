package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// panic しても 500 が返り、テストプロセス自体は落ちないことを確認する
func TestRecoverMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus int
		wantBody   string
	}{
		{
			name: "panicを回復して500を返す",
			handler: func(_ http.ResponseWriter, _ *http.Request) {
				panic("想定外のバグ")
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "サーバーエラーが発生しました",
		},
		{
			name: "ランタイムエラーのpanicも回復する",
			handler: func(_ http.ResponseWriter, _ *http.Request) {
				var s []int
				_ = s[1] // 範囲外アクセスで panic（runtime error）
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "サーバーエラーが発生しました",
		},
		{
			name: "正常なハンドラは素通しする",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
			},
			wantStatus: http.StatusOK,
			wantBody:   `"ok"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptest: 実際にサーバーを起動せずミドルウェアを検証する
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			recoverMiddleware(tt.handler).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want to contain %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

// panic の詳細（バグの内容）がクライアントに漏れないことを確認する
func TestRecoverMiddlewareHidesPanicDetail(t *testing.T) {
	secret := "内部の秘密な詳細"
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic(secret)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	recoverMiddleware(handler).ServeHTTP(rec, req)

	if strings.Contains(rec.Body.String(), secret) {
		t.Errorf("panicの内容がレスポンスに漏れている: %s", rec.Body.String())
	}
}
