package domain_test

import (
	"slices"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/domain"
)

// TestCanTransitProject は掲載状態の遷移表を固定する。
//
// 目的: 遷移ルールの一元定義を守る。壊れると「公開中の案件を下書きに戻せない」
// 「終了した案件が勝手に復活する」等、掲載フローの根幹が崩れる。
//
// 観点: **全状態 × 全遷移先（3×3=9通り）を機械的に網羅**し、期待表と突き合わせる。
// さらに「検査総数 9・許可数 4」を検算する——状態を増やしたのに期待表の更新を
// 忘れると検算が落ち、書き漏れが検出される（tsunagu-works の型）。
func TestCanTransitProject(t *testing.T) {
	// 許可される遷移の期待表（実装と独立に書く＝実装のコピーにしない）
	allowed := map[string]bool{
		"draft→published":  true, // 公開
		"published→draft":  true, // 非公開
		"published→closed": true, // 募集終了
		"closed→published": true, // 再募集
	}

	total := 0
	allowedCount := 0
	for _, from := range domain.AllProjectStatuses {
		for _, to := range domain.AllProjectStatuses {
			total++
			key := string(from) + "→" + string(to)
			got := domain.CanTransitProject(from, to)
			if got != allowed[key] {
				t.Errorf("%s: got %v, want %v", key, got, allowed[key])
			}
			if got {
				allowedCount++
			}
		}
	}

	// 検算: 状態が増えたら 9 が変わり、遷移を増やしたら 4 が変わる。
	// どちらもこのテストの期待表を見直すシグナル
	if total != 9 {
		t.Errorf("検査総数: 9 を期待したが %d（状態が増えたら期待表も更新する）", total)
	}
	if allowedCount != 4 {
		t.Errorf("許可遷移数: 4 を期待したが %d（遷移表と期待表がズレている）", allowedCount)
	}
}

// TestProjectTransitionsFrom は「ある状態から行ける先」の列挙を固定する。
// エラーメッセージ（可能な遷移の提示）と画面のボタン出し分けが依存する。
func TestProjectTransitionsFrom(t *testing.T) {
	tests := []struct {
		from domain.ProjectStatus
		want []domain.ProjectStatus
	}{
		{from: domain.ProjectDraft, want: []domain.ProjectStatus{domain.ProjectPublished}},
		{from: domain.ProjectPublished, want: []domain.ProjectStatus{domain.ProjectDraft, domain.ProjectClosed}},
		{from: domain.ProjectClosed, want: []domain.ProjectStatus{domain.ProjectPublished}},
	}
	for _, tt := range tests {
		t.Run(string(tt.from), func(t *testing.T) {
			got := domain.ProjectTransitionsFrom(tt.from)
			if !slices.Equal(got, tt.want) {
				t.Errorf("%s から: %v を期待したが %v", tt.from, tt.want, got)
			}
		})
	}
}
