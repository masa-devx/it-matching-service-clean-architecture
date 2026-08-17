package auth

import (
	"errors"
	"net/mail"
)

// 資格情報（email / password）の規則。DB に CHECK が無く、フォームの Zod は直叩きで
// 回避できるため、サーバー側 Go が唯一の防衛線。視点（company / talent）に依らない
// 認証ドメインの規則としてここで共有する
var (
	ErrInvalidEmail   = errors.New("メールアドレスの形式が正しくありません")
	ErrPasswordLength = errors.New("パスワードは8文字以上72バイト以下にしてください")
)

// 認証フローの業務エラー（company / talent 共通の語彙）
var (
	// ErrEmailTaken は email の重複（handler が 409 に変換する）
	ErrEmailTaken = errors.New("このメールアドレスは既に登録されています")
	// ErrAuthFailed は認証失敗。理由（不存在・パスワード不一致・ロール違い）は区別しない
	ErrAuthFailed = errors.New("メールアドレスまたはパスワードが正しくありません")
)

// ValidateCredentials は email の形式とパスワード長を検証する。
// len(string) はバイト数: bcrypt の上限（72バイト）に合わせてバイトで数える
func ValidateCredentials(email, password string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}
	if len(password) < 8 || len(password) > 72 {
		return ErrPasswordLength
	}
	return nil
}
