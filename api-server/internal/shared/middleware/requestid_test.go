package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/middleware"
)

// TestRequestID は request_id の採番・伝播ルールを固定する。
//
// 目的: 呼び出し元（Next.js 等）の X-Request-ID を尊重してログを突き合わせ可能にしつつ、
// 外部入力を無検証でログに流さない（ログインジェクション対策）ことを保証する。
// 壊れると「フロントとバックで ID が合わない」か「ログに改行・制御文字が混入する」。
//
// 観点: 正当な ID は尊重（context とレスポンスヘッダに同じ値）／不正な ID
// （短すぎ・長すぎ・改行や記号入り・空）は破棄して採番（16進16文字）。
func TestRequestID(t *testing.T) {
	tests := []struct {
		name     string
		incoming string
		wantKeep bool
	}{
		{name: "正当な ID（Next.js 由来の UUID 形式）は尊重", incoming: "0f47ac10-58cc-4372-a567-0e02b2c3d479", wantKeep: true},
		{name: "英数16文字（Go の採番形式）も尊重", incoming: "a1b2c3d4e5f60718", wantKeep: true},
		{name: "8文字ちょうどは尊重（境界）", incoming: "abcd1234", wantKeep: true},
		{name: "7文字は破棄（境界）", incoming: "abcd123", wantKeep: false},
		{name: "65文字は破棄（境界）", incoming: strings.Repeat("a", 65), wantKeep: false},
		{name: "改行入りは破棄（ログインジェクション）", incoming: "abcd1234\ninjected=true", wantKeep: false},
		{name: "記号入りは破棄", incoming: "abcd_1234!", wantKeep: false},
		{name: "空なら採番", incoming: "", wantKeep: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got, _ = middleware.RequestIDFrom(r.Context())
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.incoming != "" {
				req.Header.Set("X-Request-ID", tt.incoming)
			}
			rec := httptest.NewRecorder()
			middleware.RequestID(inner).ServeHTTP(rec, req)

			if tt.wantKeep {
				if got != tt.incoming {
					t.Errorf("受信 ID の尊重を期待したが: got %q", got)
				}
			} else {
				if got == tt.incoming {
					t.Errorf("不正な ID %q がそのまま使われている", tt.incoming)
				}
				if len(got) != 16 {
					t.Errorf("採番 ID は16進16文字のはずが %q", got)
				}
			}
			if header := rec.Header().Get("X-Request-ID"); header != got {
				t.Errorf("レスポンスヘッダ %q と context %q が一致しない", header, got)
			}
		})
	}
}
