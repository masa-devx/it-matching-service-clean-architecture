package validator

import (
	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
)

// 資格情報の規則は視点に依らないため shared/auth に集約し（#30）、
// エラー名は従来どおりこのパッケージからも参照できるよう再輸出する
var (
	ErrInvalidEmail   = auth.ErrInvalidEmail
	ErrPasswordLength = auth.ErrPasswordLength
)

// Signup は企業サインアップ入力を検証する。
// name の長さは DB の CHECK 制約が最終防衛線として存在するため、ここでは検証しない
func Signup(input company.TsunaguWorksCompanySignupInput) error {
	return auth.ValidateCredentials(string(input.Email), input.Password)
}
