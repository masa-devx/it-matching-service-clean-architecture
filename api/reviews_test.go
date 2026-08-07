package main

import (
	"strings"
	"testing"
)

// TestValidateReview はレビューの入力検証を網羅する。
//
// 目的: 評価の範囲外の値や空のコメントが保存されるのを防ぐこと。
// 評価は1〜5の5段階であることが機能の前提で、範囲外の値が入ると
// 平均値の計算や星の表示が壊れる。またコメントが空だと、
// 星の数字だけが残り「なぜその評価なのか」が分からない記録になる。
//
// 観点: 評価の範囲を境界の両側から（0・1・5・6）/ コメントの必須と文字数上限。
// DBのCHECK制約（rating BETWEEN 1 AND 5）と同じルールをアプリ側でも張り、
// 400として意味のあるメッセージを返せることを保証する。
func TestValidateReview(t *testing.T) {
	valid := reviewRequest{Rating: 5, Comment: "丁寧に対応いただきました"}

	// withValid は正常な値をベースに一部だけ差し替える（各ケースの差分を明確にする）
	withValid := func(mutate func(*reviewRequest)) reviewRequest {
		r := valid
		mutate(&r)
		return r
	}

	tests := []struct {
		name string
		req  reviewRequest
		want string
	}{
		{name: "正常系", req: valid, want: ""},

		// --- 評価の境界値 ---
		{
			name: "評価1は通る（下限の境界値）",
			req:  withValid(func(r *reviewRequest) { r.Rating = 1 }),
			want: "",
		},
		{
			name: "評価5は通る（上限の境界値）",
			req:  withValid(func(r *reviewRequest) { r.Rating = 5 }),
			want: "",
		},
		{
			name: "評価0（未選択のまま送信された場合）",
			req:  withValid(func(r *reviewRequest) { r.Rating = 0 }),
			want: "評価は1〜5で選択してください",
		},
		{
			name: "評価6",
			req:  withValid(func(r *reviewRequest) { r.Rating = 6 }),
			want: "評価は1〜5で選択してください",
		},

		// --- コメント ---
		{
			name: "コメントが空",
			req:  withValid(func(r *reviewRequest) { r.Comment = "" }),
			want: "コメントは必須です",
		},
		{
			name: "コメント2000文字は通る（境界値）",
			req:  withValid(func(r *reviewRequest) { r.Comment = strings.Repeat("あ", 2000) }),
			want: "",
		},
		{
			name: "コメント2001文字",
			req:  withValid(func(r *reviewRequest) { r.Comment = strings.Repeat("あ", 2001) }),
			want: "コメントは2000文字以内にしてください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateReview(tt.req); got != tt.want {
				t.Errorf("validateReview(rating=%d) = %q, want %q", tt.req.Rating, got, tt.want)
			}
		})
	}
}

// TestReviewSelectHidesUnpublishedPeerReview は取得クエリの条件を固定する。
//
// 目的: 同時公開の実効性を守ること。相手のレビューが公開前に取得できると、
// 画面に表示していなくてもAPIの直接呼び出しで読めてしまい、
// 「相手の評価を見てから自分の評価を書く」ことが可能になる。
// そうなると報復レビューを防ぐという機能の目的そのものが失われる。
//
// 観点: SQL に「公開済み または 自分が書いたもの」という条件が含まれること。
// クエリを組み替えて条件が外れたら、このテストが落ちる。
// SQLの文字列検査という素朴な方法だが、**この1行が機能の核心**なので、
// 消えたことに気づける仕組みを置いておく価値がある。
func TestReviewSelectHidesUnpublishedPeerReview(t *testing.T) {
	// 公開済み、または自分が書いたレビューだけを取る条件
	const requiredCondition = "published_at IS NOT NULL OR reviewer_role = $2"

	if !strings.Contains(reviewSelectSQL, requiredCondition) {
		t.Errorf("reviewSelectSQL に %q が含まれていない。\n"+
			"未公開の相手レビューが取得できると、同時公開の意味が失われる。\n"+
			"現在のSQL:\n%s", requiredCondition, reviewSelectSQL)
	}
}
