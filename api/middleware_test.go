package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRecoverMiddleware は panic からの復帰を検証する。
//
// 目的: ハンドラ内の想定外のバグ（nil参照・範囲外アクセスなど）が起きても、
// プロセスを落とさず 500 を返して次のリクエストを処理できる状態を保証する。
// このテスト自体がプロセス内で panic を起こすため、回復に失敗すればテストごと落ちる。
//
// 観点: 文字列 panic / ランタイムエラー由来の panic / 正常なハンドラを素通しすること
// （ミドルウェアが正常系に副作用を与えないことも同時に確認している）。
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

// TestRecoverMiddlewareHidesPanicDetail は panic の内容が外部に漏れないことを検証する。
//
// 目的: panic の値には変数の中身や内部構造が含まれうるため、クライアントには
// 定型文の 500 だけを返す仕様を固定する（詳細はログにのみ残す）。
// 「エラー時に何を返さないか」はセキュリティ要件であり、実装を変えても崩せないよう
// テストで縛っている。
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
