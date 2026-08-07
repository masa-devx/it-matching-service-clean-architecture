package main

import (
	"fmt"
	"testing"
)

// TestCanTransitionWorkReport は稼働報告の遷移ルールを「全状態 × 全遷移先 × 両ロール」で網羅する。
//
// 目的: 稼働報告を「第三者に確認された記録」として成立させること。この表が崩れると、
// 人材が自分の報告を承認できてしまい、稼働報告が単なる自己申告に落ちる。
// 稼働報告は検収と報酬計算の根拠になるため、誰が確認したかが記録の価値そのものになる。
// また approved を終端に保つことで、「承認済み＝確定した実績」という前提を守っている。
//
// 観点: 3状態 × 3遷移先 × 2ロール ＝ 18通りをすべて列挙し、
// 許可される3通り以外がすべて拒否されることを保証する。特に検証したいのは:
//   - 承認・差し戻しが企業だけに許されているか
//   - 再提出（rejected → submitted）が人材だけに許されているか
//   - approved から何も実行できないか（承認の取り消しができないこと）
//   - 同じ状態への「遷移」が拒否されるか
func TestCanTransitionWorkReport(t *testing.T) {
	// 許可される遷移だけを列挙する。ここに無い組み合わせは、
	// 下のループですべて「拒否されること」を検証する
	allowed := map[statusTransition][]string{
		{from: workReportStatusSubmitted, to: workReportStatusApproved}: {roleCompany},
		{from: workReportStatusSubmitted, to: workReportStatusRejected}: {roleCompany},
		{from: workReportStatusRejected, to: workReportStatusSubmitted}: {roleTalent},
	}

	roles := []string{roleCompany, roleTalent}
	cases := 0

	for _, from := range workReportStatuses {
		for _, to := range workReportStatuses {
			for _, role := range roles {
				cases++

				want := false
				for _, actor := range allowed[statusTransition{from: from, to: to}] {
					if actor == role {
						want = true
					}
				}

				t.Run(fmt.Sprintf("%s→%s/%s", from, to, role), func(t *testing.T) {
					if got := canTransitionWorkReport(from, to, role); got != want {
						t.Errorf("canTransitionWorkReport(%q, %q, %q) = %v, want %v",
							from, to, role, got, want)
					}
				})
			}
		}
	}

	// 状態を増やしたときに網羅が崩れていないかを検算する
	if want := len(workReportStatuses) * len(workReportStatuses) * len(roles); cases != want {
		t.Errorf("検証した組み合わせ = %d, want %d", cases, want)
	}
}

// TestWorkReportTransitionsMatchSpec は遷移表そのものの形を固定する。
//
// 目的: 上のテストは期待値をテスト側の表から導いているため、実装とテストを
// 同時に書き換えると気づけない。ここでは件数と性質だけを独立に検査し、
// 遷移を増やす変更がレビューで必ず目に入るようにする。
//
// 観点: 許可された遷移の数 / 実行者が空の行が無いこと /
// approved（承認済み）が from として現れないこと。
func TestWorkReportTransitionsMatchSpec(t *testing.T) {
	const wantTransitions = 3

	if len(workReportTransitions) != wantTransitions {
		t.Errorf("許可された遷移の数 = %d, want %d（遷移を追加・削除したらこの数も更新すること）",
			len(workReportTransitions), wantTransitions)
	}

	for transition, actors := range workReportTransitions {
		if len(actors) == 0 {
			t.Errorf("%q → %q に実行者が設定されていない", transition.from, transition.to)
		}

		// 承認済みから出る遷移があると、確定した実績が後から覆ることになる
		if transition.from == workReportStatusApproved {
			t.Errorf("承認済みの報告から %q への遷移が定義されている", transition.to)
		}
	}
}
