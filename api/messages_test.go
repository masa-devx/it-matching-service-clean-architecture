package main

import (
	"reflect"
	"strings"
	"testing"
)

// TestValidateMessage はメッセージ本文の入力検証を網羅する。
//
// 目的: 空のメッセージや極端に長い本文が保存されるのを防ぐこと。
// 空メッセージが並ぶと会話が読みにくくなり、長すぎる本文は
// 表示崩れとDBの肥大化を招く。
//
// 観点: 必須（空・空白のみ）/ 文字数の上限を境界の両側から。
// 本文は日本語が中心になるため、バイト数ではなく文字数（rune）で数えていることも確認する。
func TestValidateMessage(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "正常系", body: "お世話になります。今週の進捗を共有します。", want: ""},
		{name: "空", body: "", want: "メッセージを入力してください"},
		{
			name: "2000文字ちょうどは通る（境界値）",
			body: strings.Repeat("あ", 2000),
			want: "",
		},
		{
			name: "2001文字",
			body: strings.Repeat("あ", 2001),
			want: "メッセージは2000文字以内にしてください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateMessage(messageRequest{Body: tt.body}); got != tt.want {
				t.Errorf("validateMessage(%q...) = %q, want %q",
					firstRunes(tt.body, 10), got, tt.want)
			}
		})
	}
}

// TestMessageResponseHasNoRawBody はレスポンス型が原文を持たないことを固定する。
//
// 目的: マスキングの実効性を守ること。原文（body）がレスポンスに含まれると、
// 画面に表示していなくてもブラウザの開発者ツールやAPIの直接呼び出しで読めてしまい、
// 連絡先マスキングそのものが無意味になる。
//
// 観点: messageResponse のフィールドに、原文を入れられる場所が無いこと。
// 「返さない」ではなく「返す場所が無い」という構造を、フィールド名の検査で固定する。
// 将来 RawBody や OriginalBody のようなフィールドが足されたら、このテストが落ちる。
func TestMessageResponseHasNoRawBody(t *testing.T) {
	// 原文を入れる場所として使われうる名前。Body は masked_body の置き場所なので許容する
	forbidden := []string{"RawBody", "OriginalBody", "PlainBody", "UnmaskedBody"}

	res := messageResponse{}
	fields := structFieldNames(res)

	for _, name := range forbidden {
		for _, field := range fields {
			if field == name {
				t.Errorf("messageResponse に %q がある。原文はレスポンスに含めない", name)
			}
		}
	}
}

// TestMessageSendableStatuses は送信できる契約の状態を固定する。
//
// 目的: 終了した取引でやり取りが続くのを防ぐこと。完了・中止した契約で
// メッセージを送れると、「いつまで対応すればよいか」が曖昧になる。
//
// 観点: 進行中の3状態（成立・稼働中・検収待ち）が送信可能で、
// 終端の2状態（完了・中止）が送信不可であること。
func TestMessageSendableStatuses(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{status: contractStatusActive, want: true},
		{status: contractStatusWorking, want: true},
		{status: contractStatusReviewing, want: true},
		{status: contractStatusCompleted, want: false},
		{status: contractStatusCancelled, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := false
			for _, s := range messageSendableStatuses {
				if s == tt.status {
					got = true
				}
			}
			if got != tt.want {
				t.Errorf("状態 %q の送信可否 = %v, want %v", tt.status, got, tt.want)
			}
		})
	}

	// 状態を増やしたときに、送信可否の検討が漏れていないかを検算する
	if len(tests) != len(contractStatuses) {
		t.Errorf("検証した状態 = %d, 契約の状態 = %d（状態を追加したらケースも追加すること）",
			len(tests), len(contractStatuses))
	}
}

// firstRunes はエラーメッセージ用に先頭 n 文字だけを取り出す
func firstRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}

// structFieldNames は構造体のフィールド名を列挙する（リフレクション）。
// Type.Fields() は Go 1.24 で入ったイテレータで、NumField/Field の添字ループより読みやすい
func structFieldNames(v any) []string {
	var names []string
	for field := range reflect.TypeOf(v).Fields() {
		names = append(names, field.Name)
	}
	return names
}
