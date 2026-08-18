package domain_test

import (
	"slices"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/domain"
)

// TestCanTransitApplication は応募の遷移表を全網羅で固定する。
//
// 目的: 選考フローの業務ルール（誰が・どこから・どこへ）を実行可能な仕様として保証する。
// 表が変わればこのテストが落ちる＝遷移の追加・削除が必ずレビューの目に触れる。
//
// 観点: 2 actor × 6状態 × 6状態 = 72通りをすべて判定し、許可6件だけを期待値に列挙する。
// 特に「actor が違えば同じ from→to でも拒否される」（talent はオファーできない等）を含む。
func TestCanTransitApplication(t *testing.T) {
	allowed := map[[3]string]bool{
		{"company", "applied", "offered"}:  true,
		{"company", "applied", "rejected"}: true,
		{"talent", "applied", "withdrawn"}: true,
		{"talent", "offered", "accepted"}:  true,
		{"talent", "offered", "declined"}:  true,
		{"talent", "offered", "withdrawn"}: true,
	}

	total := 0
	allowedCount := 0
	for _, actor := range domain.AllActors {
		for _, from := range domain.AllApplicationStatuses {
			for _, to := range domain.AllApplicationStatuses {
				total++
				want := allowed[[3]string{string(actor), string(from), string(to)}]
				got := domain.CanTransitApplication(actor, from, to)
				if got != want {
					t.Errorf("CanTransitApplication(%s, %s, %s) = %v, want %v",
						actor, from, to, got, want)
				}
				if got {
					allowedCount++
				}
			}
		}
	}

	// 検算: 状態や actor を増やしたとき、期待値の検討漏れをここで検出する
	if total != 72 {
		t.Errorf("組み合わせ総数: 72 を期待したが %d（状態か actor が増えたら期待値を見直す）", total)
	}
	if allowedCount != 6 {
		t.Errorf("許可された遷移数: 6 を期待したが %d", allowedCount)
	}
}

// TestApplicationTerminalStates は終端状態を固定する。
//
// 目的: 決着した応募（accepted / rejected / withdrawn / declined）が
// どの actor によっても二度と動かせないことを保証する（決着の巻き戻しは業務上の事故）。
//
// 観点: 終端4状態 × 2 actor × 全遷移先がすべて拒否されること。
func TestApplicationTerminalStates(t *testing.T) {
	terminals := []domain.ApplicationStatus{
		domain.ApplicationAccepted,
		domain.ApplicationRejected,
		domain.ApplicationWithdrawn,
		domain.ApplicationDeclined,
	}

	for _, from := range terminals {
		for _, actor := range domain.AllActors {
			if tos := domain.ApplicationTransitionsFor(actor, from); len(tos) != 0 {
				t.Errorf("終端状態 %s から %s が遷移できてしまう: %v", from, actor, tos)
			}
		}
	}
}

// TestApplicationTransitionsFor は遷移先の列挙を固定する。
//
// 目的: エラーメッセージ・画面のボタン出し分けが使う「actor 別の可能な操作」を保証する。
// 観点: applied / offered について actor ごとの遷移先リスト（順序は AllApplicationStatuses 準拠）。
func TestApplicationTransitionsFor(t *testing.T) {
	tests := []struct {
		actor domain.Actor
		from  domain.ApplicationStatus
		want  []domain.ApplicationStatus
	}{
		{
			domain.ActorCompany, domain.ApplicationApplied,
			[]domain.ApplicationStatus{domain.ApplicationOffered, domain.ApplicationRejected},
		},
		{
			domain.ActorTalent, domain.ApplicationApplied,
			[]domain.ApplicationStatus{domain.ApplicationWithdrawn},
		},
		{domain.ActorCompany, domain.ApplicationOffered, nil},
		{
			domain.ActorTalent, domain.ApplicationOffered,
			[]domain.ApplicationStatus{domain.ApplicationAccepted, domain.ApplicationWithdrawn, domain.ApplicationDeclined},
		},
	}

	for _, tt := range tests {
		got := domain.ApplicationTransitionsFor(tt.actor, tt.from)
		if !slices.Equal(got, tt.want) {
			t.Errorf("ApplicationTransitionsFor(%s, %s) = %v, want %v",
				tt.actor, tt.from, got, tt.want)
		}
	}
}
