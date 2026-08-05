package main

import (
	"strings"
	"testing"
)

// TestValidateSignup は会員登録の入力検証を網羅する。
//
// 目的: 「登録できてはいけない入力」がすべて弾かれることを保証する。ここが緩むと
// 不正なメールアドレスのユーザーや、bcryptの72バイト制限に引っかかるパスワードが
// DBに入り、ログインできないアカウントが生まれる。
//
// 観点: email形式 / パスワードの長さ（8文字・72バイトの境界を両側から） / roleの列挙値。
// 検証結果を「エラーメッセージ文字列」で比較しているのは、文言自体がAPIの仕様
// （クライアントにそのまま表示される）だからである。
func TestValidateSignup(t *testing.T) {
	tests := []struct {
		name string
		req  signupRequest
		want string // 期待するエラーメッセージ（空文字=バリデーション通過）
	}{
		{
			name: "正常系: talent",
			req:  signupRequest{Email: "user@example.com", Password: "password123", Role: "talent"},
			want: "",
		},
		{
			name: "正常系: company",
			req:  signupRequest{Email: "user@example.com", Password: "password123", Role: "company"},
			want: "",
		},
		{
			name: "email形式不正",
			req:  signupRequest{Email: "not-an-email", Password: "password123", Role: "talent"},
			want: "メールアドレスの形式が不正です",
		},
		{
			name: "email空",
			req:  signupRequest{Email: "", Password: "password123", Role: "talent"},
			want: "メールアドレスの形式が不正です",
		},
		{
			name: "パスワード8文字未満",
			req:  signupRequest{Email: "user@example.com", Password: "short12", Role: "talent"},
			want: "パスワードは8文字以上にしてください",
		},
		{
			name: "パスワード8文字ちょうどは通過（境界値）",
			req:  signupRequest{Email: "user@example.com", Password: "12345678", Role: "talent"},
			want: "",
		},
		{
			name: "パスワード72バイト超（bcrypt上限）",
			req:  signupRequest{Email: "user@example.com", Password: strings.Repeat("a", 73), Role: "talent"},
			want: "パスワードは72文字以内にしてください",
		},
		{
			name: "パスワード72バイトちょうどは通過（境界値）",
			req:  signupRequest{Email: "user@example.com", Password: strings.Repeat("a", 72), Role: "talent"},
			want: "",
		},
		{
			name: "role不正",
			req:  signupRequest{Email: "user@example.com", Password: "password123", Role: "admin"},
			want: "role は company または talent を指定してください",
		},
		{
			name: "role空",
			req:  signupRequest{Email: "user@example.com", Password: "password123", Role: ""},
			want: "role は company または talent を指定してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateSignup(tt.req)
			if got != tt.want {
				t.Errorf("validateSignup() = %q, want %q", got, tt.want)
			}
		})
	}
}
