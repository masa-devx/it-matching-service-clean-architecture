package main

import (
	"database/sql"
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

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// loginFailMessage は「未登録email」と「パスワード違い」で必ず同一にする。
// 文言やステータスを分けると、攻撃者が登録済みemailの一覧を探れてしまう（ユーザー列挙）
const loginFailMessage = "メールアドレスまたはパスワードが正しくありません"

// dummyPasswordHash は未登録email のときに照合する捨てハッシュ。
// 照合をスキップすると応答が速くなり、時間差からアカウントの存在が推測できてしまうため、
// 実在ユーザーと同じ計算コストを常に支払う（タイミング攻撃対策）
const dummyPasswordHash = "$2a$10$0/Pf/ST0lhDyD3bAU4d/9.Lw9uG//JrhTSXZgehs2u/E17je7jxi2"

// handleLogin は POST /login。email/password を照合し JWT を返す
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "リクエストボディが不正です", err)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	var (
		userID int64
		hash   string
	)
	err := db.QueryRow(
		`SELECT id, password_hash FROM users WHERE email = $1`,
		req.Email,
	).Scan(&userID, &hash)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(req.Password))
		writeError(w, http.StatusUnauthorized, loginFailMessage, nil)
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "ログイン処理に失敗しました", err)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, loginFailMessage, nil)
		return
	}

	token, err := issueToken(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ログイン処理に失敗しました", err)
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token})
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
