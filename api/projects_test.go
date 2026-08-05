package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// TestValidateProject は案件掲載の入力検証を網羅する。
//
// 目的: 掲載データの不整合を保存前に防ぐこと。特に単価の下限・上限の逆転は
// DB側の CHECK 制約でも弾かれるが、それだけでは 500 になってしまうため、
// アプリ側で 400 と意味のあるメッセージを返せることを保証する。
//
// 観点: 必須（タイトル）/ 文字数・個数の上限 / 単価の範囲と大小関係（同額の境界含む）/
// 稼働時間の上限 / status の列挙値。
// 各ケースは withValid ヘルパーで「正常な値から1項目だけ壊す」形にしており、
// 何を検証しているのかが差分で読み取れるようにしている。
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

// TestIntQuery はクエリ文字列の整数パースとクランプを検証する。
//
// 目的: 一覧の limit に上限を設け、`?limit=100000` のような指定で全件取得されるのを防ぐこと
// （DoS の一種になりうる）。同時に、不正な値でエラーにせず既定値へフォールバックすることで
// 一覧表示そのものは壊れないようにしている。
//
// 観点: 未指定・非数値での既定値 / 上限超過と下限未満のクランプ / 負の値。
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

// TestParseProjectFilter は検索条件のパースを検証する。
//
// 目的: クエリ文字列を検索条件の構造体へ変換する際の「指定なし」の扱いを固定すること。
// ゼロ値＝絞り込まない、という規約が崩れると、条件を送っていないのに結果が減る
// （または remote_ok=false で出社案件だけになる）といった直感に反する挙動を招く。
//
// 観点: 全項目未指定 / 全条件の指定 / スキルの正規化（空白・重複）/
// remote_ok=false は絞り込まないこと / 数値でない値を指定なし扱いにすること。
func TestParseProjectFilter(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  projectFilter
	}{
		{
			name:  "指定なしはゼロ値（絞り込まない）",
			query: "",
			want:  projectFilter{Skills: []string{}},
		},
		{
			name:  "全条件",
			query: "?skills=Go,React&rate_min=4000&rate_max=8000&hours_max=20&remote_ok=true&q=API",
			want: projectFilter{
				Skills: []string{"Go", "React"}, RateMin: 4000, RateMax: 8000,
				HoursMax: 20, RemoteOnly: true, Keyword: "API",
			},
		},
		{
			name:  "スキルの空白と重複は正規化される",
			query: "?skills=%20Go%20,React,Go,",
			want:  projectFilter{Skills: []string{"Go", "React"}},
		},
		{
			name:  "remote_ok=false は絞り込まない",
			query: "?remote_ok=false",
			want:  projectFilter{Skills: []string{}},
		},
		{
			name:  "数値でない値は指定なし扱い",
			query: "?rate_min=abc",
			want:  projectFilter{Skills: []string{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/projects"+tt.query, nil)
			got := parseProjectFilter(r)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseProjectFilter() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestValidateFilter は検索条件の妥当性検証を確認する。
//
// 目的: 明らかに矛盾した条件（下限 > 上限）や過大な指定を 400 で返すこと。
// 黙って丸めると「指定した条件と違う結果が返る」ため、利用者が誤解する。
//
// 観点: 条件なしが通ること / スキル個数・キーワード長・単価・稼働時間の上限 /
// 単価は両方指定されたときだけ大小関係を見る（片方だけの指定は有効）。
func TestValidateFilter(t *testing.T) {
	tests := []struct {
		name string
		f    projectFilter
		want string
	}{
		{name: "正常系", f: projectFilter{Skills: []string{"Go"}, RateMin: 4000, RateMax: 8000}, want: ""},
		{name: "条件なしも通る", f: projectFilter{}, want: ""},
		{
			name: "スキル11個",
			f:    projectFilter{Skills: make([]string, 11)},
			want: "スキルは10個以内で指定してください",
		},
		{
			name: "キーワード101文字",
			f:    projectFilter{Keyword: strings.Repeat("あ", 101)},
			want: "キーワードは100文字以内で指定してください",
		},
		{
			name: "単価の下限が上限を超える",
			f:    projectFilter{RateMin: 9000, RateMax: 3000},
			want: "単価の下限は上限以下で指定してください",
		},
		{
			name: "下限のみの指定は通る（上限は未指定）",
			f:    projectFilter{RateMin: 9000},
			want: "",
		},
		{
			name: "稼働169時間",
			f:    projectFilter{HoursMax: 169},
			want: "週の稼働時間は0〜168の範囲で指定してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateFilter(tt.f); got != tt.want {
				t.Errorf("validateFilter() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestBuildProjectWhereUsesPlaceholders は動的に組み立てた WHERE 句の安全性を検証する。
//
// 目的: 検索条件を SQL に組み立てる処理で、値が文字列連結されていないことを保証する。
// 動的クエリは SQL インジェクションの温床になりやすく、実装を変えたときに
// うっかり fmt.Sprintf で値を埋め込んでも気づけるよう、テストで縛っている。
//
// 観点: SQL 文字列に値（スキル名・数値・インジェクション文字列）が現れないこと /
// プレースホルダに渡す引数の個数が条件の数と一致すること
// （remote_ok は値を取らない条件なので引数が増えない点も含めて固定している）。
func TestBuildProjectWhereUsesPlaceholders(t *testing.T) {
	f := projectFilter{
		Skills: []string{"Go"}, RateMin: 4000, RateMax: 8000,
		HoursMax: 20, RemoteOnly: true, Keyword: "' OR 1=1--",
	}
	where, args := buildProjectWhere(f)

	// SQL文字列に値そのものが現れてはいけない（現れる＝文字列連結している）
	for _, leaked := range []string{"Go", "4000", "8000", "20", "OR 1=1"} {
		if strings.Contains(where, leaked) {
			t.Errorf("値がSQLに埋め込まれている: %q in %q", leaked, where)
		}
	}
	// status + skills + rate_min + rate_max + hours_max + keyword = 6個（remote_ok は値を取らない）
	if len(args) != 6 {
		t.Errorf("args = %d件, want 6件: %+v", len(args), args)
	}
}
