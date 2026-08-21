package validator_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/validator"
)

// TestOfferMessage はオファーメッセージの長さ検証を固定する。
//
// 目的: 上限500「文字」の防衛線はこの validator だけ（DB に CHECK なし・フォームは curl で回避可能）。
// 壊れると巨大なメッセージが保存され、一覧レスポンスの肥大や表示崩れにつながる。
//
// 観点: 境界を両側から（500/501文字）。「バイト数ではなく文字数」であることを
// マルチバイトのケースで固定する（500文字の日本語 = 1500バイトが通ること）。
// 空文字は「メッセージ無し」として正当。
func TestOfferMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{name: "空はメッセージ無しとして正当", message: "", wantErr: false},
		{name: "500文字ちょうどは通る（境界）", message: strings.Repeat("a", 500), wantErr: false},
		{name: "501文字は拒否（境界）", message: strings.Repeat("a", 501), wantErr: true},
		{
			// 「あ」は UTF-8 で3バイト: 500文字 = 1500バイト。バイト数で数える実装ミスだと拒否してしまう
			name:    "マルチバイト500文字（1500バイト）は通る",
			message: strings.Repeat("あ", 500),
			wantErr: false,
		},
		{name: "マルチバイト501文字は拒否", message: strings.Repeat("あ", 501), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.OfferMessage(tt.message)
			if tt.wantErr && !errors.Is(err, validator.ErrOfferMessageTooLong) {
				t.Errorf("ErrOfferMessageTooLong を期待したが: %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("許可されるべき入力が拒否された: %v", err)
			}
		})
	}
}
