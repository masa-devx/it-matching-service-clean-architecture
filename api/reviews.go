package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type reviewRequest struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

// reviewResponse は1件のレビュー。
//
// 未公開のレビューはそもそも取得しない（SELECT で除外する）ため、
// この型に「公開されているか」のフラグは持たない。
// 返ってきた時点で「自分のもの」か「公開済み」のどちらかだと確定している
type reviewResponse struct {
	ID           int64      `json:"id"`
	ReviewerRole string     `json:"reviewer_role"`
	Rating       int        `json:"rating"`
	Comment      string     `json:"comment"`
	SubmittedAt  time.Time  `json:"submitted_at"`
	PublishedAt  *time.Time `json:"published_at"`
}

// reviewListResponse は契約のレビュー一覧。
//
// published は「双方のレビューが公開されたか」。まだなら相手のレビューは
// 配列に含まれない。画面はこのフラグを見て「相手の提出を待っています」と伝える
// （空配列だけ返すと「相手が書いていない」のか「見せてもらえない」のか区別できない）
type reviewListResponse struct {
	Reviews   []reviewResponse `json:"reviews"`
	Published bool             `json:"published"`
	// 自分がすでに提出したか。画面が投稿フォームを出すかどうかの判断に使う
	Submitted bool `json:"submitted"`
}

// handleCreateReview は POST /contracts/{id}/reviews。契約の当事者のみ。
//
// 完了した契約にだけ投稿できる。稼働中や検収待ちの段階で評価を書けると、
// 「悪い評価をつけられたくないから検収を通す」といった圧力が生まれる
func handleCreateReview(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", nil)
		return
	}

	contractID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(r.Context(), w, http.StatusNotFound, "契約が見つかりません", nil)
		return
	}

	var req reviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "リクエストボディが不正です", err)
		return
	}
	req.Comment = strings.TrimSpace(req.Comment)
	if msg := validateReview(req); msg != "" {
		writeError(r.Context(), w, http.StatusBadRequest, msg, nil)
		return
	}

	role, err := fetchRole(userID)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "レビューの投稿に失敗しました", err)
		return
	}

	// 所有チェック（当事者か）と状態チェック（完了しているか）を1クエリで済ませる
	status, err := fetchOwnedContractStatus(contractID, userID, role)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(r.Context(), w, http.StatusNotFound, "契約が見つかりません", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "レビューの投稿に失敗しました", err)
		return
	}
	if status != contractStatusCompleted {
		writeError(r.Context(), w, http.StatusConflict, "完了した契約にのみレビューを投稿できます", nil)
		return
	}

	published, err := insertReview(contractID, role, req)
	// 二重投稿は事前SELECTではなくUNIQUE違反で検知する（TOCTOU を避ける）
	if isUniqueViolation(err) {
		writeError(r.Context(), w, http.StatusConflict, "この契約にはすでにレビューを投稿しています", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "レビューの投稿に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]bool{"published": published})
}

// insertReview はレビューを保存し、両者が揃っていれば同時に公開する。
// 戻り値は「このレビューによって公開されたか」。
//
// 【同時公開の実装】
// 保存と公開判定を1つのトランザクションで行う。Go 側で「相手も提出済みか」を
// 判定してから UPDATE する形にすると、判定と更新の間に相手が提出した場合に
// 「両方とも相手を未提出と判断して、どちらも公開されない」状態が起こりうる（TOCTOU）。
//
// トランザクション内でSQLとして数えれば、その隙間ができない。
// また片方だけ公開される状態も構造的に作れない（同じ UPDATE で両方を更新するため）
func insertReview(contractID int64, role string, req reviewRequest) (bool, error) {
	tx, err := db.Begin()
	if err != nil {
		return false, err
	}
	// 先に defer しておけば、途中で return してもロールバックされる。
	// Commit 済みなら Rollback は何もしない
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(
		`INSERT INTO reviews (contract_id, reviewer_role, rating, comment)
		 VALUES ($1, $2, $3, $4)`,
		contractID, role, req.Rating, req.Comment,
	); err != nil {
		return false, err
	}

	// 自分の分を入れたあとで数える。2件あれば双方が提出済み＝公開の条件が揃った
	var count int
	if err := tx.QueryRow(
		`SELECT count(*) FROM reviews WHERE contract_id = $1`,
		contractID,
	).Scan(&count); err != nil {
		return false, err
	}

	published := count == 2
	if published {
		// 両方の行に同じ時刻を入れる。1つの UPDATE で更新することで、
		// 「片方だけ公開されている」状態が存在しえない
		if _, err := tx.Exec(
			`UPDATE reviews SET published_at = now()
			 WHERE contract_id = $1 AND published_at IS NULL`,
			contractID,
		); err != nil {
			return false, err
		}
	}

	return published, tx.Commit()
}

// 公開済みのレビューと、自分が書いたレビューだけを取得する。
//
// 「相手のレビューを画面で隠す」のではなく、そもそも取得しない。
// 返してしまうと、画面に表示していなくてもブラウザの開発者ツールや
// APIの直接呼び出しで読めてしまい、同時公開の意味が失われる（#117 と同じ原則）
const reviewSelectSQL = `
	SELECT id, reviewer_role, rating, comment, submitted_at, published_at
	FROM reviews
	WHERE contract_id = $1
	  AND (published_at IS NOT NULL OR reviewer_role = $2)
	ORDER BY reviewer_role`

// handleListReviews は GET /contracts/{id}/reviews。契約の当事者のみ。
func handleListReviews(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", nil)
		return
	}

	contractID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(r.Context(), w, http.StatusNotFound, "契約が見つかりません", nil)
		return
	}

	role, err := fetchRole(userID)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "レビューの取得に失敗しました", err)
		return
	}

	// 所有チェックは契約に対して行う（当事者でなければレビューも見られない）
	if _, err := fetchOwnedContractStatus(contractID, userID, role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(r.Context(), w, http.StatusNotFound, "契約が見つかりません", nil)
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "レビューの取得に失敗しました", err)
		return
	}

	rows, err := db.Query(reviewSelectSQL, contractID, role)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "レビューの取得に失敗しました", err)
		return
	}
	defer func() { _ = rows.Close() }()

	res := reviewListResponse{Reviews: make([]reviewResponse, 0, 2)}
	for rows.Next() {
		var (
			review      reviewResponse
			publishedAt sql.NullTime
		)
		if err := rows.Scan(&review.ID, &review.ReviewerRole, &review.Rating,
			&review.Comment, &review.SubmittedAt, &publishedAt); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "レビューの取得に失敗しました", err)
			return
		}
		review.PublishedAt = nullTimePtr(publishedAt)

		if publishedAt.Valid {
			res.Published = true
		}
		if review.ReviewerRole == role {
			res.Submitted = true
		}
		res.Reviews = append(res.Reviews, review)
	}
	if err := rows.Err(); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "レビューの取得に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// validateReview はレビューの入力検証
func validateReview(req reviewRequest) string {
	// DBのCHECK制約と同じルールをアプリ側でも検証し、意味のあるメッセージを返す
	if req.Rating < 1 || req.Rating > 5 {
		return "評価は1〜5で選択してください"
	}
	if req.Comment == "" {
		return "コメントは必須です"
	}
	if len([]rune(req.Comment)) > 2000 {
		return "コメントは2000文字以内にしてください"
	}
	return ""
}
