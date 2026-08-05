package main

import (
	"reflect"
	"strings"
	"testing"
)

// TestValidateCompanyProfile は企業プロフィールの入力検証を網羅する。
//
// 目的: 必須項目（会社名）の欠落と、文字数上限の超過を確実に弾くこと。
// 会社名は案件掲載時に表示される情報のため、空のまま保存されると一覧の見た目が壊れる。
//
// 観点: 必須チェック / 100文字・2000文字の境界（日本語で数えるため rune 単位）/
// 任意項目が空でも通ること。
func TestValidateCompanyProfile(t *testing.T) {
	tests := []struct {
		name string
		p    companyProfile
		want string
	}{
		{
			name: "正常系",
			p:    companyProfile{Name: "つなぐ株式会社", Description: "受託開発", Industry: "IT", Size: "11-50"},
			want: "",
		},
		{
			name: "会社名以外は空でも通る",
			p:    companyProfile{Name: "つなぐ株式会社"},
			want: "",
		},
		{
			name: "会社名が空",
			p:    companyProfile{Name: ""},
			want: "会社名は必須です",
		},
		{
			name: "会社名100文字ちょうどは通る（境界値）",
			p:    companyProfile{Name: strings.Repeat("あ", 100)},
			want: "",
		},
		{
			name: "会社名101文字",
			p:    companyProfile{Name: strings.Repeat("あ", 101)},
			want: "会社名は100文字以内にしてください",
		},
		{
			name: "説明2001文字",
			p:    companyProfile{Name: "つなぐ株式会社", Description: strings.Repeat("あ", 2001)},
			want: "会社説明は2000文字以内にしてください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateCompanyProfile(tt.p); got != tt.want {
				t.Errorf("validateCompanyProfile() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestValidateTalentProfile は人材プロフィールの入力検証を網羅する。
//
// 目的: 数値項目が現実にありえない値で保存されるのを防ぐ。特に週の稼働時間は
// 168時間（24h×7日）が物理的な上限であり、ドメイン知識をバリデーションに落とし込んでいる。
//
// 観点: 全項目が未入力（ゼロ値）でも通ること / 文字数・個数の上限 /
// 数値の範囲（0・70・168 などの境界を両側から）。
func TestValidateTalentProfile(t *testing.T) {
	tests := []struct {
		name string
		p    talentProfile
		want string
	}{
		{
			name: "正常系",
			p: talentProfile{
				Bio: "Goが得意です", Skills: []string{"Go", "React"},
				YearsOfExp: 5, AvailableHoursPerWeek: 20, DesiredHourlyRate: 5000, RemoteOK: true,
			},
			want: "",
		},
		{
			name: "全て未入力（ゼロ値）でも通る",
			p:    talentProfile{},
			want: "",
		},
		{
			name: "自己紹介2001文字",
			p:    talentProfile{Bio: strings.Repeat("あ", 2001)},
			want: "自己紹介は2000文字以内にしてください",
		},
		{
			name: "スキル31個",
			p:    talentProfile{Skills: make([]string, 31)},
			want: "スキルは30個以内にしてください",
		},
		{
			name: "スキル1つが51文字",
			p:    talentProfile{Skills: []string{strings.Repeat("a", 51)}},
			want: "各スキルは50文字以内にしてください",
		},
		{
			name: "経験年数が負",
			p:    talentProfile{YearsOfExp: -1},
			want: "経験年数は0〜70の範囲で入力してください",
		},
		{
			name: "経験年数70は通る（境界値）",
			p:    talentProfile{YearsOfExp: 70},
			want: "",
		},
		{
			name: "週の稼働168は通る（境界値・1週間の総時間）",
			p:    talentProfile{AvailableHoursPerWeek: 168},
			want: "",
		},
		{
			name: "週の稼働169",
			p:    talentProfile{AvailableHoursPerWeek: 169},
			want: "週の稼働可能時間は0〜168の範囲で入力してください",
		},
		{
			name: "希望時給が負",
			p:    talentProfile{DesiredHourlyRate: -1},
			want: "希望時給は0〜1000000の範囲で入力してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateTalentProfile(tt.p); got != tt.want {
				t.Errorf("validateTalentProfile() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestNormalizeSkills はスキル配列の正規化を検証する。
//
// 目的: フォームから届く「揺れた入力」（前後の空白・空欄・重複）を保存前に整えること。
// ここが甘いと DB に " Go" と "Go" が別物として入り、GINインデックスによる
// 検索（skills @> '{Go}'）で取りこぼしが起きる。
//
// 観点: trim / 空要素の除去 / 重複排除（最初の出現を残す）/
// nil を空スライスに正規化すること（JSONで null ではなく [] を返す契約のため）。
func TestNormalizeSkills(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "前後の空白を除去",
			input: []string{" Go ", "React "},
			want:  []string{"Go", "React"},
		},
		{
			name:  "空要素を除去",
			input: []string{"Go", "", "   ", "React"},
			want:  []string{"Go", "React"},
		},
		{
			name:  "重複を除去（最初の出現を残す）",
			input: []string{"Go", "React", "Go"},
			want:  []string{"Go", "React"},
		},
		{
			name:  "nil は空スライスになる（JSONで null でなく [] を返すため）",
			input: nil,
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSkills(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalizeSkills(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
