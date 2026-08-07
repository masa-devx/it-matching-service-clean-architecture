package main

import (
	"strings"
	"testing"
)

// TestMaskContactsDetects は連絡先が確実に伏せられることを固定する。
//
// 目的: 直接取引の防止＝このサービスの事業モデルそのものを守ること。
// ここが漏れると、プラットフォーム外で契約され、手数料が入らないだけでなく
// トラブル時の保護（エスクロー・稼働報告・レビュー）がすべて効かなくなる。
//
// 観点: メール・電話・URLの代表的な表記を網羅する。特に電話番号は
// ハイフンあり/なし/スペース区切り/国番号と表記が多様なため、実際に使われる形を列挙する。
// また、文の途中にある場合・複数含まれる場合も検証する
// （文頭にしかマッチしない正規表現になっていないことの確認）。
func TestMaskContactsDetects(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		// --- メールアドレス ---
		{name: "メール", body: "連絡先は foo@example.com です"},
		{name: "メール（記号入り）", body: "test.user+tag@example.co.jp までお願いします"},

		// --- 電話番号 ---
		{name: "携帯（ハイフンあり）", body: "090-1234-5678 に電話ください"},
		{name: "携帯（ハイフンなし）", body: "09012345678 です"},
		{name: "携帯（スペース区切り）", body: "090 1234 5678 まで"},
		{name: "固定電話", body: "03-1234-5678 が会社の番号です"},
		{name: "国番号", body: "+81-90-1234-5678 に連絡してください"},

		// --- URL ---
		{name: "https", body: "こちらを見てください https://example.com/contact"},
		{name: "http", body: "http://example.com のフォームから"},

		// --- 位置・複数 ---
		{name: "文頭", body: "foo@example.com に送ってください"},
		{name: "文末", body: "私のアドレスは foo@example.com"},
		{name: "複数（メール2つ）", body: "a@example.com か b@example.com へ"},
		{name: "複数（種類が違う）", body: "090-1234-5678 か foo@example.com へ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masked, wasMasked := maskContacts(tt.body)

			if !wasMasked {
				t.Errorf("maskContacts(%q) が伏せ字を入れていない", tt.body)
			}
			if !strings.Contains(masked, maskedPlaceholder) {
				t.Errorf("maskContacts(%q) = %q, 伏せ字が含まれていない", tt.body, masked)
			}
		})
	}
}

// TestMaskContactsDoesNotOverMask は正常な文章が伏せられないことを固定する。
//
// 目的: 誤検知によって正常なやり取りが阻害されるのを防ぐこと。
// 検知漏れと誤検知は両立しないため、このプロジェクトでは**誤検知を減らす方向に倒す**
// 判断をしている。日時・期間・日付・金額はメッセージに頻出するため、
// これらが伏せられると業務のやり取りそのものが成立しなくなる。
//
// 観点: 数字やハイフンを含むが電話番号ではない表記を列挙する。
// 特に「2026-08-07」（日付）は桁数だけ見ると電話番号に近く、
// 「0始まり」という条件で除外していることを確認する。
func TestMaskContactsDoesNotOverMask(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "時間帯", body: "10時から18時まで対応できます"},
		{name: "期間（ハイフン）", body: "1-2週間で完了する見込みです"},
		{name: "日付（ハイフン区切り）", body: "2026-08-07 に納品します"},
		{name: "金額の範囲", body: "単価は5000-8000円を想定しています"},
		{name: "稼働時間", body: "週20時間、合計80時間の稼働です"},
		{name: "バージョン番号", body: "Go 1.26.5 と Next.js 16 を使っています"},
		{name: "普通の文章", body: "今週の作業は認証APIの実装とテスト作成です"},
		{name: "スキーム無しのドメイン", body: "詳細は example.com の資料をご覧ください"},
		{name: "箇条書きの番号", body: "1. 設計 2. 実装 3. テスト の順で進めます"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masked, wasMasked := maskContacts(tt.body)

			if wasMasked {
				t.Errorf("maskContacts(%q) = %q, 正常な文章が伏せられた", tt.body, masked)
			}
			if masked != tt.body {
				t.Errorf("maskContacts(%q) = %q, 本文が変わっている", tt.body, masked)
			}
		})
	}
}

// TestMaskContactsKeepsSurroundingText は伏せ字以外の本文が残ることを固定する。
//
// 目的: 連絡先だけを伏せ、前後の文脈は読めるようにすること。
// 本文ごと消してしまうと、何のやり取りだったのか分からなくなり、
// 「連絡先を書いた」という事実だけが残って会話が成立しなくなる。
//
// 観点: 伏せ字の前後の文字列が保持されているか。
func TestMaskContactsKeepsSurroundingText(t *testing.T) {
	body := "お世話になります。連絡先は foo@example.com です。よろしくお願いします。"

	masked, wasMasked := maskContacts(body)

	if !wasMasked {
		t.Fatalf("maskContacts(%q) が伏せ字を入れていない", body)
	}
	for _, keep := range []string{"お世話になります。", "連絡先は", "よろしくお願いします。"} {
		if !strings.Contains(masked, keep) {
			t.Errorf("maskContacts(%q) = %q, %q が失われている", body, masked, keep)
		}
	}
	if strings.Contains(masked, "foo@example.com") {
		t.Errorf("maskContacts(%q) = %q, 連絡先が残っている", body, masked)
	}
}

// TestLooksLikePhoneNumber は桁数による判定を直接検証する。
//
// 目的: 電話番号の判定基準（日本の番号は0始まりで10〜11桁）を明示的に固定すること。
// maskContacts 経由のテストでは「正規表現で候補にならなかった」のか
// 「候補にはなったが桁数で除外された」のかが区別できないため、判定関数を単体で検証する。
//
// 観点: 10桁（固定電話）・11桁（携帯）の境界と、その前後（9桁・12桁）。
func TestLooksLikePhoneNumber(t *testing.T) {
	tests := []struct {
		name      string
		candidate string
		want      bool
	}{
		{name: "10桁（固定電話・境界値）", candidate: "03-1234-5678", want: true},
		{name: "11桁（携帯・境界値）", candidate: "090-1234-5678", want: true},
		{name: "9桁（短すぎる）", candidate: "03-1234-567", want: false},
		{name: "12桁（長すぎる）", candidate: "090-1234-56789", want: false},
		{name: "国番号（+81 + 10桁）", candidate: "+81-90-1234-5678", want: true},
		{name: "4桁（時間帯など）", candidate: "1018", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := looksLikePhoneNumber(tt.candidate); got != tt.want {
				t.Errorf("looksLikePhoneNumber(%q) = %v, want %v", tt.candidate, got, tt.want)
			}
		})
	}
}
