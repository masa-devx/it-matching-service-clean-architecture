package main

import "regexp"

// maskedPlaceholder は伏せ字。何が起きたかが伝わる文言にする
// （「****」だけだと、システムの不具合と区別がつかない）
const maskedPlaceholder = "[連絡先は表示できません]"

// 連絡先とみなすパターン。
//
// 【この機能の目標設定】
// 完璧な検出は不可能である。「ゼロキュウゼロのイチニイ…」と書かれたら検出できないし、
// 全角・記号の混入・分割送信など、抜け道はいくらでもある。
// 目指すのは「簡単にはできなくする」こと——コピペで連絡先を送れないようにすることで、
// 大半の直接取引を防ぐ。抜け道を探す相手には、規約と原文の保存（監査）で対処する。
//
// 【トレードオフの向き】
// 検知漏れ（連絡先が通る）と誤検知（正常な文が伏せられる）は両立しない。
// **誤検知を減らす方向に倒す**——「10時から18時」や「1-2週間」が伏せられて
// 正常なやり取りが阻害されるほうが、実害が大きいため。
var (
	// メールアドレス。@ を挟む記号の並びは日常の文章にほぼ出ないため、
	// 比較的安全に検出できる
	emailPattern = regexp.MustCompile(`[\w.+-]+@[\w-]+\.[\w.-]+`)

	// 電話番号。日本の番号は 0 始まりで10〜11桁、または +81 形式。
	//
	// 誤検知を避けるための条件が3つある:
	//  1. 先頭が 0 または +（「2026-08-07」のような日付は 2 始まりなので対象外）
	//  2. 数字の合計が10〜11桁（「1-2週間」は2桁、「10時から18時」も4桁なので対象外）
	//  3. 前後が数字や英字でない（長い数字列の一部を切り取らない）
	//
	// 区切りはハイフン・スペース・括弧を許容する（090-1234-5678 / 090 1234 5678 / (090)1234-5678）
	phonePattern = regexp.MustCompile(
		`(?:\+81[-\s]?|0)\d{1,4}[-\s(）)]?\d{1,4}[-\s(）)]?\d{3,4}`)

	// URL。スキーム付きのものだけを対象にする。
	// 「詳細は example.com を参照」のようなスキームなしの表記まで拾うと
	// 正常な文章が伏せられるため、あえて検出しない（誤検知を減らす方針）
	urlPattern = regexp.MustCompile(`https?://[^\s、。]+`)
)

// maskContacts はメッセージ本文から連絡先とみられる箇所を伏せ字に置き換える。
// 伏せ字が入ったかどうかも返す（UIで「一部を伏せました」と理由を伝えるため）。
//
// 保存時に一度だけ呼ぶこと。表示のたびに実行すると、正規表現を改善したときに
// 過去のメッセージの見え方まで変わり、「あのとき相手に何が見えていたか」を
// 再現できなくなる（紛争時に争えなくなる）
func maskContacts(body string) (string, bool) {
	masked := body

	// URL を先に処理する。URLの中にメールアドレスらしき文字列や数字列が
	// 含まれることがあるため（例: https://example.com/user@id/09012345678）、
	// 先に丸ごと伏せてしまうほうが結果が安定する
	masked = urlPattern.ReplaceAllString(masked, maskedPlaceholder)
	masked = emailPattern.ReplaceAllString(masked, maskedPlaceholder)
	masked = maskPhoneNumbers(masked)

	return masked, masked != body
}

// maskPhoneNumbers は電話番号らしき箇所だけを伏せる。
//
// 正規表現だけでは「数字の合計桁数」を表現できないため、
// 候補を抽出したうえで桁数を数えて判定する。
// 正規表現を複雑にするより、抽出と判定を分けたほうが読めるし、
// 「なぜこの文字列は電話番号ではないのか」をテストで説明しやすい
func maskPhoneNumbers(body string) string {
	return phonePattern.ReplaceAllStringFunc(body, func(candidate string) string {
		if !looksLikePhoneNumber(candidate) {
			// 電話番号でなければ元の文字列をそのまま返す（伏せない）
			return candidate
		}
		return maskedPlaceholder
	})
}

// looksLikePhoneNumber は候補文字列が日本の電話番号の桁数を満たすかを返す。
//
// 日本の電話番号は 0 始まりで10桁（固定電話・03-1234-5678）か
// 11桁（携帯・090-1234-5678）。国番号表記（+81）は先頭の 0 が省かれるため9〜10桁になる
func looksLikePhoneNumber(candidate string) bool {
	digits := 0
	for _, r := range candidate {
		if r >= '0' && r <= '9' {
			digits++
		}
	}

	// +81 形式は先頭の 0 が省略される分、1桁少なくなる
	if len(candidate) > 0 && candidate[0] == '+' {
		return digits >= 11 && digits <= 12 // 81 + 9〜10桁
	}
	return digits >= 10 && digits <= 11
}
