package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type workReportRequest struct {
	// 報告対象の週。週内のどの日を送ってもよく、サーバー側で月曜に丸める
	// （クライアントごとに週の計算が違う可能性があるため、揃えるのはサーバーの責務）
	WeekStart string `json:"week_start"`
	Hours     int    `json:"hours"`
	Summary   string `json:"summary"`
}

type workReportResponse struct {
	ID         int64  `json:"id"`
	ContractID int64  `json:"contract_id"`
	WeekStart  string `json:"week_start"`
	Hours      int    `json:"hours"`
	Summary    string `json:"summary"`
	Status     string `json:"status"`
	// 差し戻しの理由。承認時・未確認時は空文字
	ReviewNote  string     `json:"review_note"`
	SubmittedAt time.Time  `json:"submitted_at"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
}

// handleCreateWorkReport は POST /contracts/{id}/work-reports。人材のみ。
//
// 契約が稼働中（working）のときだけ提出できる。成立直後（active）や検収待ち（reviewing）に
// 報告を足せると、「いつの作業か」が契約の進行と食い違ってしまうため
func handleCreateWorkReport(w http.ResponseWriter, r *http.Request) {
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

	var req workReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "リクエストボディが不正です", err)
		return
	}
	if msg := validateWorkReport(req); msg != "" {
		writeError(r.Context(), w, http.StatusBadRequest, msg, nil)
		return
	}

	// 所有チェック（自分の契約か）と状態チェック（稼働中か）を1クエリで済ませる。
	// 他人の契約なら1行も返らないので、存在しないものとして404にする
	var status string
	err = db.QueryRow(
		`SELECT c.status FROM contracts c
		 JOIN talents t ON t.id = c.talent_id
		 WHERE c.id = $1 AND t.user_id = $2`,
		contractID, userID,
	).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(r.Context(), w, http.StatusNotFound, "契約が見つかりません", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "稼働報告の提出に失敗しました", err)
		return
	}
	if status != contractStatusWorking {
		// 409: 入力も権限も正しいが、契約の状態が提出を許さない
		writeError(r.Context(), w, http.StatusConflict, "稼働中の契約にのみ報告を提出できます", nil)
		return
	}

	res, err := insertWorkReport(contractID, req)
	// 重複は事前SELECTではなくUNIQUE違反で検知する。
	// 「確認してから挿入」は確認と挿入の間に別リクエストが入ると破れるため（TOCTOU）
	if isUniqueViolation(err) {
		writeError(r.Context(), w, http.StatusConflict, "その週の報告はすでに提出されています", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "稼働報告の提出に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusCreated, res)
}

// insertWorkReport は報告を1件作る。
//
// 週の丸めは date_trunc('week', ...) をDB側で行う。PostgreSQL の week は
// ISO週（月曜始まり）なので、週内のどの日を渡しても必ず同じ月曜に揃う。
// クライアントの週計算を信用せず、保存する値をサーバー側で確定させる
func insertWorkReport(contractID int64, req workReportRequest) (workReportResponse, error) {
	var (
		res        workReportResponse
		weekStart  time.Time
		reviewedAt sql.NullTime
	)
	err := db.QueryRow(
		`INSERT INTO work_reports (contract_id, week_start, hours, summary)
		 VALUES ($1, date_trunc('week', $2::date)::date, $3, $4)
		 RETURNING id, contract_id, week_start, hours, summary, status, review_note,
		           submitted_at, reviewed_at`,
		contractID, req.WeekStart, req.Hours, req.Summary,
	).Scan(&res.ID, &res.ContractID, &weekStart, &res.Hours, &res.Summary,
		&res.Status, &res.ReviewNote, &res.SubmittedAt, &reviewedAt)
	if err != nil {
		return res, err
	}
	// DATE は時刻部分を持たないため、日付だけの文字列に整形して返す
	// （クライアントが時刻やタイムゾーンを気にせず扱えるようにする）
	res.WeekStart = weekStart.Format("2006-01-02")
	res.ReviewedAt = nullTimePtr(reviewedAt)
	return res, nil
}

// handleUpdateWorkReport は PUT /work-reports/{id}。人材のみ。
//
// 差し戻された報告を修正し、同時に提出済みへ戻す。
//
// #97（案件）では「内容の編集」と「状態の変更」をエンドポイントごと分けたが、
// 稼働報告では分けない。案件には「下書き」という状態があり、保存と公開が
// 別の意思決定だったのに対し、稼働報告に下書きは無く、
// 内容を出すこと自体が提出だから（提出しない報告は存在しない）。
// 分けると「直す → 出し直す」の2手になり、操作と意味がずれる
func handleUpdateWorkReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", nil)
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(r.Context(), w, http.StatusNotFound, "稼働報告が見つかりません", nil)
		return
	}

	var req workReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "リクエストボディが不正です", err)
		return
	}
	// 週は変更できないため（対象週を変えるなら別の報告になる）、この経路では検証しない
	if msg := validateWorkReportContent(req); msg != "" {
		writeError(r.Context(), w, http.StatusBadRequest, msg, nil)
		return
	}

	current, err := fetchOwnedWorkReportStatus(id, userID, roleTalent)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(r.Context(), w, http.StatusNotFound, "稼働報告が見つかりません", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "稼働報告の更新に失敗しました", err)
		return
	}

	// 再提出は rejected → submitted の遷移。遷移表で判定することで、
	// 「承認済みの報告を書き換える」「提出済みを何度も出し直す」を同じ仕組みで防ぐ
	if !canTransitionWorkReport(current, workReportStatusSubmitted, roleTalent) {
		writeError(r.Context(), w, http.StatusConflict, "差し戻された報告のみ修正できます", nil)
		return
	}

	res, err := resubmitWorkReport(id, current, req)
	if errors.Is(err, sql.ErrNoRows) {
		// 読んでから更新するまでの間に企業が操作した（楽観ロックの検知）
		writeError(r.Context(), w, http.StatusConflict, "他の操作により状態が変わりました。画面を更新してください", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "稼働報告の更新に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

// resubmitWorkReport は差し戻された報告を修正して再提出する。
//
// submitted_at を更新するのは「いつ出された報告か」として最新の提出時刻が意味を持つため。
// review_note（差し戻し理由）は空に戻す——修正済みの報告に古い指摘が残っていると、
// 企業が「まだ直っていない」と誤読する
func resubmitWorkReport(id int64, from string, req workReportRequest) (workReportResponse, error) {
	var (
		res        workReportResponse
		weekStart  time.Time
		reviewedAt sql.NullTime
	)
	err := db.QueryRow(
		`UPDATE work_reports
		 SET hours = $1, summary = $2,
		     status = $3, review_note = '', reviewed_at = NULL,
		     submitted_at = now(), updated_at = now()
		 WHERE id = $4 AND status = $5
		 RETURNING id, contract_id, week_start, hours, summary, status, review_note,
		           submitted_at, reviewed_at`,
		req.Hours, req.Summary, workReportStatusSubmitted, id, from,
	).Scan(&res.ID, &res.ContractID, &weekStart, &res.Hours, &res.Summary,
		&res.Status, &res.ReviewNote, &res.SubmittedAt, &reviewedAt)
	if err != nil {
		return res, err
	}
	res.WeekStart = weekStart.Format("2006-01-02")
	res.ReviewedAt = nullTimePtr(reviewedAt)
	return res, nil
}

// fetchOwnedWorkReportStatus は自分が当事者である報告の現在の状態を返す。
//
// 報告は契約にぶら下がるため、所有の確認は契約の当事者かどうかで行う。
// 所有チェックを WHERE に埋め込むことで、当事者でない報告は1行も返らない
func fetchOwnedWorkReportStatus(id, userID int64, role string) (string, error) {
	var query string
	switch role {
	case roleCompany:
		query = `SELECT wr.status FROM work_reports wr
		         JOIN contracts c  ON c.id  = wr.contract_id
		         JOIN companies co ON co.id = c.company_id
		         WHERE wr.id = $1 AND co.user_id = $2`
	case roleTalent:
		query = `SELECT wr.status FROM work_reports wr
		         JOIN contracts c ON c.id = wr.contract_id
		         JOIN talents   t ON t.id = c.talent_id
		         WHERE wr.id = $1 AND t.user_id = $2`
	default:
		return "", sql.ErrNoRows
	}

	var status string
	err := db.QueryRow(query, id, userID).Scan(&status)
	return status, err
}

// validateWorkReport は新規提出の検証。週の指定を含む
func validateWorkReport(req workReportRequest) string {
	if req.WeekStart == "" {
		return "対象の週は必須です"
	}
	// 日付として解釈できることをここで確認する。DBに任せると
	// パースエラーが500として返ってしまい、入力ミスだと分からない
	if _, err := time.Parse("2006-01-02", req.WeekStart); err != nil {
		return "対象の週は YYYY-MM-DD 形式で指定してください"
	}
	return validateWorkReportContent(req)
}

// validateWorkReportContent は週を除く内容の検証。再提出（週は変えられない）でも使う
func validateWorkReportContent(req workReportRequest) string {
	// DBのCHECK制約と同じルールをアプリ側でも検証し、意味のあるメッセージを返す
	if req.Hours < 0 || req.Hours > 168 {
		return "稼働時間は0〜168の範囲で入力してください"
	}
	if req.Summary == "" {
		return "作業内容は必須です"
	}
	if len([]rune(req.Summary)) > 2000 {
		return "作業内容は2000文字以内にしてください"
	}
	return ""
}
