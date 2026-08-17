package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword は平文パスワードを bcrypt でハッシュ化する。
// bcrypt は salt を内蔵し（同じ平文でも毎回違うハッシュ）、意図的に遅い（総当たり耐性）
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("パスワードのハッシュ化に失敗: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword はハッシュと平文を照合する（一致しなければ error）
func VerifyPassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}

// ユーザー不存在時にも bcrypt を1回実行し、応答時間の差で email の存在有無を
// 漏らさないためのダミーハッシュ（同一401の3点セット: 文言・ステータス・応答時間）
var dummyHash string

func init() {
	h, err := HashPassword("timing-equalizer-dummy")
	if err != nil {
		panic(fmt.Sprintf("ダミーハッシュの生成に失敗: %v", err))
	}
	dummyHash = h
}

// VerifyPasswordWithDummy はダミーハッシュとの照合で bcrypt 1回分の時間を消費する。
// ログイン時にユーザーが存在しない場合に呼ぶ
func VerifyPasswordWithDummy(password string) {
	_ = VerifyPassword(dummyHash, password)
}
