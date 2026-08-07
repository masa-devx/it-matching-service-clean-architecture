package main

import "testing"

// TestCanTransitionProject は掲載状態の遷移ルールを全組み合わせで固定する。
//
// 目的: 掲載状態が意図しない経路で変わるのを防ぐこと。この表が崩れると、
// 「一度も公開していない案件が募集終了になる」「終了した案件が下書きに戻る」といった
// 説明のつかない状態が生まれ、案件一覧の意味そのものが壊れる。
// また企業が意図せず案件を公開してしまう事故にも直結する。
//
// 観点: 取りうる状態3つ × 遷移先3つ ＝ 9通りをすべて列挙し、
// 許可される4通り以外がすべて拒否されることを保証する（ホワイトリストであることの確認）。
// 同じ状態への「変更」（draft → draft 等）も禁止に含めている。
// 表にケースを足し忘れると網羅が崩れるため、最後にケース数の検算も行う。
func TestCanTransitionProject(t *testing.T) {
	tests := []struct {
		name string
		from string
		to   string
		want bool
	}{
		// --- 許可される4通り ---
		{name: "下書き→公開（公開する）", from: projectStatusDraft, to: projectStatusPublished, want: true},
		{name: "公開→下書き（非公開に戻す）", from: projectStatusPublished, to: projectStatusDraft, want: true},
		{name: "公開→終了（募集を終了する）", from: projectStatusPublished, to: projectStatusClosed, want: true},
		{name: "終了→公開（再募集する）", from: projectStatusClosed, to: projectStatusPublished, want: true},

		// --- 意味が不明瞭なため禁止 ---
		{name: "下書き→終了（一度も公開していない案件は終了できない）", from: projectStatusDraft, to: projectStatusClosed, want: false},
		{name: "終了→下書き（終了した案件は下書きに戻せない）", from: projectStatusClosed, to: projectStatusDraft, want: false},

		// --- 同じ状態への変更は遷移ではない ---
		{name: "下書き→下書き", from: projectStatusDraft, to: projectStatusDraft, want: false},
		{name: "公開→公開", from: projectStatusPublished, to: projectStatusPublished, want: false},
		{name: "終了→終了", from: projectStatusClosed, to: projectStatusClosed, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canTransitionProject(tt.from, tt.to); got != tt.want {
				t.Errorf("canTransitionProject(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}

	// 状態を増やしたときにケースの追加を忘れると、網羅していないのに通ってしまう。
	// 想定の組み合わせ数と一致することを検算しておく
	if want := len(projectStatuses) * len(projectStatuses); len(tests) != want {
		t.Errorf("ケース数 = %d, want %d（状態を追加したらテストケースも追加すること）", len(tests), want)
	}
}

// TestProjectTransitionsAreWhitelist は遷移表がホワイトリストとして機能していることを確かめる。
//
// 目的: 「禁止を書き忘れる」事故を防ぐこと。禁止リスト方式だと状態を追加したときに
// 禁止の記述を漏らし、意図せず遷移可能になる。許可だけを列挙する方式であれば、
// 表に無い組み合わせは自動的に拒否される。
//
// 観点: 表に登録されている遷移の数が想定どおり（4通り）であること。
// 遷移を足すときは、この数字も更新する＝レビューで気づける形にしている。
func TestProjectTransitionsAreWhitelist(t *testing.T) {
	const wantAllowed = 4

	if len(projectTransitions) != wantAllowed {
		t.Errorf("許可された遷移の数 = %d, want %d（遷移を追加・削除したらこの数も更新すること）",
			len(projectTransitions), wantAllowed)
	}

	// 表の値が false の項目があると「登録されているのに不許可」という紛らわしい状態になる。
	// 許可するものだけを true で登録する運用を固定する
	for transition, allowed := range projectTransitions {
		if !allowed {
			t.Errorf("%q → %q が false で登録されている。許可しない遷移は表から削除すること",
				transition.from, transition.to)
		}
	}
}
