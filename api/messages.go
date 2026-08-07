package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
)

type messageRequest struct {
	Body string `json:"body"`
}

// messageResponse は表示用のメッセージ。
//
// フィールドは Body ひとつだけで、ここには masked_body（伏せ字を入れた文字列）が入る。
// 原文用のフィールドを作らないのは、「返さない」ではなく「返す場所が無い」状態にするため
// （#97 で status を SET 句に書かなかったのと同じ守り方。うっかり詰め込む余地を消す）
type messageResponse struct {
	ID         int64  `json:"id"`
	SenderRole string `json:"sender_role"`
	// 送信者の表示名。誰の発言かを画面で示すために返す
	SenderName string    `json:"sender_name"`
	Body       string    `json:"body"`
	Masked     bool      `json:"masked"`
	CreatedAt  time.Time `json:"created_at"`
}

type messageListResponse struct {
	Messages []messageResponse `json:"messages"`
	Total    int               `json:"total"`
}

// messageSendableStatuses はメッセージを送信できる契約の状態。
//
// 完了・中止した契約では送れない（記録の閲覧はできる）。
// 取引が終わったあともやり取りが続くと「いつまで対応すればよいか」が曖昧になるため、
// 取引の終了と同時にやり取りも終える。相手への評価はレビューで伝えられる。
//
// 稼働報告が working のときだけ提出できたのとは条件が違う。
// 条件の相談は稼働前（active）にも起きるため、進行中ならいつでも送れるようにしている
var messageSendableStatuses = []string{
	contractStatusActive,
	contractStatusWorking,
	contractStatusReviewing,
}

// handleCreateMessage は POST /contracts/{id}/messages。契約の当事者のみ。
//
// 送信者のロールはリクエストで受け取らず、検証済みトークンから引いたロールを使う
// （クライアント供給値を信用しない＝なりすまし防止）
func handleCreateMessage(w http.ResponseWriter, r *http.Request) {
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

	var req messageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "リクエストボディが不正です", err)
		return
	}
	req.Body = strings.TrimSpace(req.Body)
	if msg := validateMessage(req); msg != "" {
		writeError(r.Context(), w, http.StatusBadRequest, msg, nil)
		return
	}

	role, err := fetchRole(userID)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "メッセージの送信に失敗しました", err)
		return
	}

	// 所有チェック（当事者か）と状態チェック（進行中か）を1クエリで済ませる。
	// 当事者でなければ1行も返らないので、存在しないものとして404にする
	status, err := fetchOwnedContractStatus(contractID, userID, role)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(r.Context(), w, http.StatusNotFound, "契約が見つかりません", nil)
		return
	}
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "メッセージの送信に失敗しました", err)
		return
	}
	if !slices.Contains(messageSendableStatuses, status) {
		// 409: 入力も権限も正しいが、契約の状態が送信を許さない
		writeError(r.Context(), w, http.StatusConflict, "終了した契約にはメッセージを送信できません", nil)
		return
	}

	// マスキングは保存時に一度だけ行う。表示のたびに計算すると、正規表現を改善したときに
	// 過去のメッセージの見え方まで変わり、「あのとき相手に何が見えていたか」を再現できなくなる
	maskedBody, wasMasked := maskContacts(req.Body)

	res, err := insertMessage(contractID, role, req.Body, maskedBody, wasMasked)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "メッセージの送信に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusCreated, res)
}

// insertMessage は原文と表示用の両方を保存し、表示用だけを返す。
//
// RETURNING で masked_body を Body に受けているのがポイント。
// 原文（body）は保存されるが、この関数から外へは出ない
func insertMessage(contractID int64, senderRole, body, maskedBody string, masked bool) (messageResponse, error) {
	res := messageResponse{SenderRole: senderRole}
	err := db.QueryRow(
		`INSERT INTO messages (contract_id, sender_role, body, masked_body, masked)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, masked_body, masked, created_at`,
		contractID, senderRole, body, maskedBody, masked,
	).Scan(&res.ID, &res.Body, &res.Masked, &res.CreatedAt)
	return res, err
}

// 会話の取得。masked_body だけを SELECT し、body（原文）は取得しない。
//
// 「返さない」のではなく「取らない」ことで、レスポンスへ詰め込む経路そのものを断つ。
// 送信者の表示名は、ロールに応じて companies / talents のどちらかから引く
// （CASE で選ぶことで、結合は1回ずつで済む）
const messageSelectSQL = `
	SELECT m.id, m.sender_role,
	       CASE WHEN m.sender_role = 'company' THEN co.name ELSE t.display_name END,
	       m.masked_body, m.masked, m.created_at
	FROM messages m
	JOIN contracts  c  ON c.id  = m.contract_id
	JOIN companies  co ON co.id = c.company_id
	JOIN talents    t  ON t.id  = c.talent_id
	WHERE m.contract_id = $1
	ORDER BY m.created_at, m.id`

// handleListMessages は GET /contracts/{id}/messages。契約の当事者のみ。
//
// 会話は当事者が同じものを見る記録なので、視点で内容を変えない（契約・稼働報告と同じ判断）。
// 並びは古い順——会話は上から下へ読むため（稼働報告が新しい週から並ぶのとは逆）。
//
// ページネーションは設けない。1契約のやり取りは限られており、
// 途中で切れると会話の流れが追えなくなる（増えて実害が出たら検討する）
func handleListMessages(w http.ResponseWriter, r *http.Request) {
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
		writeError(r.Context(), w, http.StatusInternalServerError, "メッセージの取得に失敗しました", err)
		return
	}

	// 所有チェックは契約に対して行う（契約の当事者でなければ会話も見られない）
	if _, err := fetchOwnedContractStatus(contractID, userID, role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(r.Context(), w, http.StatusNotFound, "契約が見つかりません", nil)
			return
		}
		writeError(r.Context(), w, http.StatusInternalServerError, "メッセージの取得に失敗しました", err)
		return
	}

	rows, err := db.Query(messageSelectSQL, contractID)
	if err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "メッセージの取得に失敗しました", err)
		return
	}
	defer func() { _ = rows.Close() }()

	messages := make([]messageResponse, 0)
	for rows.Next() {
		var m messageResponse
		if err := rows.Scan(&m.ID, &m.SenderRole, &m.SenderName,
			&m.Body, &m.Masked, &m.CreatedAt); err != nil {
			writeError(r.Context(), w, http.StatusInternalServerError, "メッセージの取得に失敗しました", err)
			return
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		writeError(r.Context(), w, http.StatusInternalServerError, "メッセージの取得に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusOK, messageListResponse{
		Messages: messages, Total: len(messages),
	})
}

// validateMessage はメッセージの入力検証
func validateMessage(req messageRequest) string {
	if req.Body == "" {
		return "メッセージを入力してください"
	}
	if len([]rune(req.Body)) > 2000 {
		return "メッセージは2000文字以内にしてください"
	}
	return ""
}
