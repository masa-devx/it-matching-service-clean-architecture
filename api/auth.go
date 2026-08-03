package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type signupResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// uniqueViolation は PostgreSQL の一意制約違反（SQLSTATE）
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const uniqueViolation = "23505"

// handleSignup は POST /signup。重複チェックは SELECT で事前確認せず、
// INSERT の一意制約違反を検出して 409 に変換する（事前確認方式は
// 同時リクエストのすき間で重複が生まれるため、制約こそが唯一の防波堤）
func handleSignup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "リクエストボディが不正です", err)
		return
	}

	// email は大文字小文字を区別しないのが実態なので、正規化してから保存する
	// （Foo@example.com と foo@example.com を別ユーザーにしない）
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if msg := validateSignup(req); msg != "" {
		writeError(w, http.StatusBadRequest, msg, nil)
		return
	}

	// bcrypt はソルト生成・埋め込みまで自動で行う。コストは公式推奨の既定値(10)
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "登録処理に失敗しました", err)
		return
	}

	var res signupResponse
	err = db.QueryRow(
		`INSERT INTO users (email, password_hash, role)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, role, created_at`,
		req.Email, string(hash), req.Role,
	).Scan(&res.ID, &res.Email, &res.Role, &res.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			writeError(w, http.StatusConflict, "このメールアドレスは既に登録されています", err)
			return
		}
		writeError(w, http.StatusInternalServerError, "登録処理に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusCreated, res)
}

// validateSignup は入力検証。問題があればユーザー向けメッセージを返す（なければ空文字）
func validateSignup(req signupRequest) string {
	addr, err := mail.ParseAddress(req.Email)
	if err != nil || addr.Address != req.Email {
		return "メールアドレスの形式が不正です"
	}
	if len(req.Password) < 8 {
		return "パスワードは8文字以上にしてください"
	}
	// bcrypt は72バイトを超える入力を切り捨てるため、超過は明示的に拒否する
	if len(req.Password) > 72 {
		return "パスワードは72文字以内にしてください"
	}
	if req.Role != "company" && req.Role != "talent" {
		return "role は company または talent を指定してください"
	}
	return ""
}
