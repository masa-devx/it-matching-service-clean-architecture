package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateProject(t *testing.T) {
	valid := projectRequest{
		Title:          "Go APIの開発支援",
		Description:    "認証基盤の実装",
		RequiredSkills: []string{"Go", "PostgreSQL"},
		HourlyRateMin:  4000,
		HourlyRateMax:  6000,
		HoursPerWeek:   20,
		Status:         projectStatusDraft,
	}

	// withValid は正常な値をベースに一部だけ差し替える（各ケースの差分を明確にする）
	withValid := func(mutate func(*projectRequest)) projectRequest {
		p := valid
		mutate(&p)
		return p
	}

	tests := []struct {
		name string
		req  projectRequest
		want string
	}{
		{name: "正常系", req: valid, want: ""},
		{
			name: "タイトルが空",
			req:  withValid(func(p *projectRequest) { p.Title = "" }),
			want: "案件タイトルは必須です",
		},
		{
			name: "タイトル100文字は通る（境界値）",
			req:  withValid(func(p *projectRequest) { p.Title = strings.Repeat("あ", 100) }),
			want: "",
		},
		{
			name: "タイトル101文字",
			req:  withValid(func(p *projectRequest) { p.Title = strings.Repeat("あ", 101) }),
			want: "案件タイトルは100文字以内にしてください",
		},
		{
			name: "説明5001文字",
			req:  withValid(func(p *projectRequest) { p.Description = strings.Repeat("あ", 5001) }),
			want: "案件内容は5000文字以内にしてください",
		},
		{
			name: "スキル31個",
			req:  withValid(func(p *projectRequest) { p.RequiredSkills = make([]string, 31) }),
			want: "必須スキルは30個以内にしてください",
		},
		{
			name: "単価の下限が上限を超える",
			req: withValid(func(p *projectRequest) {
				p.HourlyRateMin, p.HourlyRateMax = 9000, 3000
			}),
			want: "単価の下限は上限以下にしてください",
		},
		{
			name: "単価が同額は通る（境界値）",
			req: withValid(func(p *projectRequest) {
				p.HourlyRateMin, p.HourlyRateMax = 5000, 5000
			}),
			want: "",
		},
		{
			name: "単価が負",
			req:  withValid(func(p *projectRequest) { p.HourlyRateMin = -1 }),
			want: "単価は0〜1000000の範囲で入力してください",
		},
		{
			name: "週の稼働168は通る（境界値）",
			req:  withValid(func(p *projectRequest) { p.HoursPerWeek = 168 }),
			want: "",
		},
		{
			name: "週の稼働169",
			req:  withValid(func(p *projectRequest) { p.HoursPerWeek = 169 }),
			want: "週の稼働時間は0〜168の範囲で入力してください",
		},
		{
			name: "不正なstatus",
			req:  withValid(func(p *projectRequest) { p.Status = "open" }),
			want: "status は draft / published / closed のいずれかを指定してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateProject(tt.req); got != tt.want {
				t.Errorf("validateProject() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIntQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  int
	}{
		{name: "未指定は既定値", query: "", want: 20},
		{name: "数値でなければ既定値", query: "?limit=abc", want: 20},
		{name: "正常な値", query: "?limit=50", want: 50},
		{name: "上限を超えたらクランプ", query: "?limit=100000", want: 100},
		{name: "下限を下回ってもクランプ", query: "?limit=0", want: 1},
		{name: "負の値もクランプ", query: "?limit=-10", want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/projects"+tt.query, nil)
			if got := intQuery(r, "limit", defaultLimit, 1, maxLimit); got != tt.want {
				t.Errorf("intQuery() = %d, want %d", got, tt.want)
			}
		})
	}
}
