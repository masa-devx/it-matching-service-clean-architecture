package main

import (
	"reflect"
	"strings"
	"testing"
)

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
