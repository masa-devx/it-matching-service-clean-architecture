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
