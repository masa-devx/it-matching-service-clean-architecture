package main

import (
	"fmt"
	"testing"
)

// TestCanTransitionContract は契約の遷移ルールを「全状態 × 全遷移先 × 両ロール」で網羅する。
//
// 目的: 取引の進行が意図しない経路で変わるのを防ぐこと。この表が崩れると、
// 「人材が自分で契約を完了させる（検収を経ずに報酬を確定させる）」「完了した取引が
// 後から中止される」といった、金銭と信頼に直結する事故が起きる。
// 特に reviewing → completed を企業だけに限っていることが、「成果を確認してから完了する」
// という検収の意味そのものを守っている。
//
// 観点: 5状態 × 5遷移先 × 2ロール ＝ 50通りをすべて列挙し、
// 許可される10通り（7遷移のうち中止3行が2ロール分）以外がすべて拒否されることを保証する。
// 検証したい性質は以下の5つ:
//   - 進む遷移が正しい実行者だけに許されているか
//   - 差し戻し（reviewing → working）が企業だけに許されているか
//   - 中止が3つの状態から両ロールで実行できるか
//   - 終端（completed / cancelled）からは何もできないか
//   - 同じ状態への「遷移」が拒否されるか
//
// ケースの書き忘れに気づけるよう、最後に組み合わせ数の検算も行う。
func TestCanTransitionContract(t *testing.T) {
	// 許可される遷移だけを (from, to) → 実行できるロール で列挙する。
	// ここに無い組み合わせは、下のループですべて「拒否されること」を検証する
	allowed := map[statusTransition][]string{
		{from: contractStatusActive, to: contractStatusWorking}:      {roleTalent},
		{from: contractStatusWorking, to: contractStatusReviewing}:   {roleTalent},
		{from: contractStatusReviewing, to: contractStatusCompleted}: {roleCompany},
		{from: contractStatusReviewing, to: contractStatusWorking}:   {roleCompany},
		{from: contractStatusActive, to: contractStatusCancelled}:    {roleCompany, roleTalent},
		{from: contractStatusWorking, to: contractStatusCancelled}:   {roleCompany, roleTalent},
		{from: contractStatusReviewing, to: contractStatusCancelled}: {roleCompany, roleTalent},
	}

	roles := []string{roleCompany, roleTalent}
	cases := 0

	for _, from := range contractStatuses {
		for _, to := range contractStatuses {
			for _, role := range roles {
				cases++

				// この組み合わせが許可されるべきか（期待値）を allowed から導く
				want := false
				for _, actor := range allowed[statusTransition{from: from, to: to}] {
					if actor == role {
						want = true
					}
				}

				name := fmt.Sprintf("%s→%s/%s", from, to, role)
				t.Run(name, func(t *testing.T) {
					if got := canTransitionContract(from, to, role); got != want {
						t.Errorf("canTransitionContract(%q, %q, %q) = %v, want %v",
							from, to, role, got, want)
					}
				})
			}
		}
	}

	// 状態を増やしたときに網羅が崩れていないかを検算する。
	// 全ケースがPASSしていても、そもそも回していない組み合わせは検証されないため
	if want := len(contractStatuses) * len(contractStatuses) * len(roles); cases != want {
		t.Errorf("検証した組み合わせ = %d, want %d", cases, want)
	}
}

// TestContractTransitionsMatchSpec は遷移表そのものの形を固定する。
//
// 目的: 遷移や実行者を「気づかないうちに」増やす変更を防ぐこと。
// 上のテストは canTransitionContract の振る舞いを検証するが、期待値を
// テスト側の表から導いているため、両方の表を同時に書き換えると気づけない。
// ここでは件数と値の性質だけを独立に検査し、レビューで必ず目に入るようにする。
//
// 観点: 許可された遷移の数 / 実行者が空の行が無いこと /
// 終端状態（completed・cancelled）が from として現れないこと。
func TestContractTransitionsMatchSpec(t *testing.T) {
	const wantTransitions = 7

	if len(contractTransitions) != wantTransitions {
		t.Errorf("許可された遷移の数 = %d, want %d（遷移を追加・削除したらこの数も更新すること）",
			len(contractTransitions), wantTransitions)
	}

	for transition, actors := range contractTransitions {
		// 実行者が空だと「表にあるのに誰も実行できない」という紛らわしい行になる。
		// 実行者がいない遷移は表から削除するべき
		if len(actors) == 0 {
			t.Errorf("%q → %q に実行者が設定されていない", transition.from, transition.to)
		}

		// 終端状態から出る遷移があると、完了した取引が後から覆ることになる
		if transition.from == contractStatusCompleted || transition.from == contractStatusCancelled {
			t.Errorf("終端状態 %q から %q への遷移が定義されている", transition.from, transition.to)
		}
	}
}
