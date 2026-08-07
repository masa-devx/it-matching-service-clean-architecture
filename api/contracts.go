package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// createContractFromApplication は承諾された応募から契約を作る。
//
// 呼び出し側と同じトランザクション（tx）で実行すること。応募の承諾と契約の作成は
// 「合意が成立した」という1つの事実の裏表であり、片方だけ成功する状態を作ってはいけない
// （人材は承諾したつもりなのに契約が無い、という手で直すしかない状態になる）。
//
// 案件の条件は参照ではなく値でコピーする。案件は掲載後に編集できる（PUT /projects/{id}）ため、
// 参照のままだと契約成立後に単価を書き換えられてしまう
func createContractFromApplication(tx *sql.Tx, applicationID int64) error {
	_, err := tx.Exec(
		`INSERT INTO contracts
		   (application_id, project_id, company_id, talent_id,
		    title, hourly_rate, hours_per_week, remote_ok)
		 SELECT a.id, p.id, p.company_id, a.talent_id,
		        p.title, p.hourly_rate_max, p.hours_per_week, p.remote_ok
		 FROM applications a
		 JOIN projects p ON p.id = a.project_id
		 WHERE a.id = $1`,
		applicationID,
	)
	if err != nil {
		return fmt.Errorf("契約の作成に失敗: %w", err)
	}
	return nil
}

type contractStatusRequest struct {
	Status string `json:"status"`
}

// handleUpdateContractStatus は PATCH /contracts/{id}/status。
//
// 企業・人材のどちらも当事者なので requireRole は付けない。
// 誰がどの遷移を実行できるかは遷移表が判定する（応募の状態更新と同じ形）。
//
// 認可は3層で確認する: 認証（requireAuth）→ 所有（取得クエリのJOINに自分のIDを入れる）
// → 遷移可否（canTransitionContract）
func handleUpdateContractStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFrom(r.Context())
	if !ok {
		writeError(r.Context(), w, http.StatusUnauthorized, "認証が必要です", nil)
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(r.Context(), w, http.StatusNotFound, "契約が見つかりません", nil)
		return
	}

	var req contractStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "リクエストボディが不正です", err)
		return
	}

	// ロールはトークンではなくDBを参照する（requireRole と同じ理由：発行後の変更を反映するため）
	role, err := fetchRole(userID)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "契約の更新に失敗しました", err)
		return
	}

	current, err := fetchOwnedContractStatus(id, userID, role)
	if errors.Is(err, sql.ErrNoRows) {
		// 当事者でない契約は「存在しない」として扱う（403だと他社の取引の存在が漏れる）
		writeError(r.Context(), w, http.StatusNotFound, "契約が見つかりません", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "契約の更新に失敗しました", err)
		return
	}

	if !canTransitionContract(current, req.Status, role) {
		// 409: 入力も権限も正しいが、現在の状態からは実行できない
		writeError(r.Context(), w, http.StatusConflict, "この契約に対してその操作はできません", nil)
		return
	}

	status, err := updateContractStatus(id, current, req.Status)
	if errors.Is(err, sql.ErrNoRows) {
		// 状態を読んでから更新するまでの間に相手が操作した（楽観ロックの検知）
		writeError(r.Context(), w, http.StatusConflict, "他の操作により状態が変わりました。画面を更新してください", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "契約の更新に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": status})
}

// fetchOwnedContractStatus は自分が当事者である契約の現在の状態を返す。
//
// 契約は企業・人材の両方が当事者なので、ロールによって「自分のもの」の意味が変わる。
// 所有チェックを WHERE に埋め込むことで、当事者でない契約はそもそも1行も返らない
// （取得してからアプリで判定する形にすると、判定漏れがそのまま権限昇格になる）
func fetchOwnedContractStatus(id, userID int64, role string) (string, error) {
	var query string
	switch role {
	case roleCompany:
		query = `SELECT c.status FROM contracts c
		         JOIN companies co ON co.id = c.company_id
		         WHERE c.id = $1 AND co.user_id = $2`
	case roleTalent:
		query = `SELECT c.status FROM contracts c
		         JOIN talents t ON t.id = c.talent_id
		         WHERE c.id = $1 AND t.user_id = $2`
	default:
		return "", sql.ErrNoRows
	}

	var status string
	err := db.QueryRow(query, id, userID).Scan(&status)
	return status, err
}

// updateContractStatus は契約の状態を進める。
//
// WHERE に読み取り時の状態(from)を含めることで、読んでから更新するまでの間に
// 相手が操作していた場合は0行更新となり sql.ErrNoRows で検知できる（楽観ロック）。
// 契約は当事者が2人いるぶん、応募よりも競合が起きやすい
// （企業が検収しようとした瞬間に人材が中止する、など）。
//
// 時刻の扱いに注意:
//   - started_at は「初めて稼働に入った時刻」。差し戻し（reviewing → working）で
//     再び working になるが、そこで上書きすると開始日が後ろにずれてしまうため、
//     COALESCE で既存値を優先する（NULL のときだけ now() が入る）
//   - completed_at は completed に入ったときだけ記録する
func updateContractStatus(id int64, from, to string) (string, error) {
	var status string
	err := db.QueryRow(
		`UPDATE contracts
		 SET status = $1,
		     updated_at = now(),
		     started_at = CASE WHEN $1 = 'working' THEN COALESCE(started_at, now()) ELSE started_at END,
		     completed_at = CASE WHEN $1 = 'completed' THEN now() ELSE completed_at END
		 WHERE id = $2 AND status = $3
		 RETURNING status`,
		to, id, from,
	).Scan(&status)
	return status, err
}
