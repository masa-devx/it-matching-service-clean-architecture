package validator

import (
	"errors"
	"net/mail"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
)

// パスワードと email は DB に CHECK が無く（保存されるのはハッシュ）、フォームの Zod は
// 直叩き（curl）で回避できる。サーバー側 Go が唯一の防衛線のため、ここで検証する
// （「validator は相関ルールだけ」の原則の例外＝「他に防衛線が無い検証」も担う・#29）
var (
	ErrInvalidEmail   = errors.New("メールアドレスの形式が正しくありません")
	ErrPasswordLength = errors.New("パスワードは8文字以上72バイト以下にしてください")
)

// Signup は企業サインアップ入力を検証する。
// name の長さは DB の CHECK 制約が最終防衛線として存在するため、ここでは検証しない
func Signup(input company.TsunaguWorksCompanySignupInput) error {
	if _, err := mail.ParseAddress(string(input.Email)); err != nil {
		return ErrInvalidEmail
	}
	// len(string) はバイト数: bcrypt の上限（72バイト）に合わせてバイトで数える
	if len(input.Password) < 8 || len(input.Password) > 72 {
		return ErrPasswordLength
	}
	return nil
}
