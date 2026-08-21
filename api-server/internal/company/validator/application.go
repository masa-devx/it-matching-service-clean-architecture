package validator

import (
	"errors"
	"unicode/utf8"
)

// ErrOfferMessageTooLong はオファーメッセージの長さ超過（利用者にそのまま見せる文言）
var ErrOfferMessageTooLong = errors.New("オファーメッセージは500文字以内で入力してください")

// maxOfferMessageRunes はオファーメッセージの上限。
// パスワード（bcrypt の72バイト＝技術都合）と違い「人が読む文章の長さ」という業務ルールなので、
// バイト数ではなく利用者の感覚と一致する文字数（rune）で数える
const maxOfferMessageRunes = 500

// OfferMessage はオファーメッセージを検証する（空は「メッセージ無し」で正当）
func OfferMessage(message string) error {
	if utf8.RuneCountInString(message) > maxOfferMessageRunes {
		return ErrOfferMessageTooLong
	}
	return nil
}
