package validator_test

import (
	"errors"
	"strings"
	"testing"

	openapi_types "github.com/oapi-codegen/runtime/types"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/validator"
)

// TestSignup はサインアップ入力の検証を固定する。
//
// 目的: パスワード・email はサーバー側 Go が唯一の防衛線（DB は CHECK なし・フォーム Zod は
// curl で回避可能）。壊れると弱いパスワードや不正な email で登録できてしまう。
//
// 観点: パスワード長の境界を両側から（7/8・72/73 バイト）。
// 「文字数」ではなく「バイト数」で数えること（bcrypt の 72 バイト上限）を
// マルチバイト文字のケースで固定する。email は形式の基本ケースのみ（RFC 完全準拠は狙わない）。
func TestSignup(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{name: "正常系", email: "user@example.com", password: "password123", wantErr: nil},
		{name: "email に @ が無い", email: "not-an-email", password: "password123", wantErr: validator.ErrInvalidEmail},
		{name: "パスワード7バイトは拒否（境界）", email: "user@example.com", password: "1234567", wantErr: validator.ErrPasswordLength},
		{name: "パスワード8バイトは通る（境界）", email: "user@example.com", password: "12345678", wantErr: nil},
		{name: "パスワード72バイトは通る（境界）", email: "user@example.com", password: strings.Repeat("a", 72), wantErr: nil},
		{name: "パスワード73バイトは拒否（境界）", email: "user@example.com", password: strings.Repeat("a", 73), wantErr: validator.ErrPasswordLength},
		{
			// 「あ」は UTF-8 で3バイト: 25文字=75バイト。文字数で数える実装ミスだと通ってしまう
			name:     "マルチバイト25文字（75バイト）は拒否",
			email:    "user@example.com",
			password: strings.Repeat("あ", 25),
			wantErr:  validator.ErrPasswordLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := company.TsunaguWorksCompanySignupInput{
				Email:    openapi_types.Email(tt.email),
				Password: tt.password,
				Name:     "テスト株式会社",
			}

			err := validator.Signup(input)

			if tt.wantErr == nil && err != nil {
				t.Errorf("許可されるべき入力が拒否された: %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("%v を期待したが: %v", tt.wantErr, err)
			}
		})
	}
}
