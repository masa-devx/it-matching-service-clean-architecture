// Package validator は talent API の入力検証を置く。
// 検証の役割分担は company 側と同じ: 単一フィールド制約は仕様＋DB CHECK が守り、
// ここには「仕様で表現できない検証」と「他に防衛線が無い検証」だけを書く
package validator

import (
	talent "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/talent"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
)

// 資格情報の規則は shared/auth に集約されている（#30）。エラー名を再輸出する
var (
	ErrInvalidEmail   = auth.ErrInvalidEmail
	ErrPasswordLength = auth.ErrPasswordLength
)

// Signup は人材サインアップ入力を検証する。
// display_name の長さは DB の CHECK 制約（1〜50）が最終防衛線として存在するため検証しない
func Signup(input talent.TsunaguWorksTalentSignupInput) error {
	return auth.ValidateCredentials(string(input.Email), input.Password)
}
