package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// tokenTTL はトークンの有効期限。漏えい時の被害を時間で限定する
const tokenTTL = 24 * time.Hour

// jwtSecret は署名鍵。initJWT で起動時に一度だけ読み込む
var jwtSecret []byte

// initJWT は署名鍵を環境変数から読み込む。
// 未設定のまま起動すると全トークンが検証不能な事故になるため、起動を止める（fail fast）
func initJWT() error {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return fmt.Errorf("JWT_SECRET が設定されていません（.env を確認してください）")
	}
	jwtSecret = []byte(secret)
	return nil
}

// issueToken はユーザーIDを sub に持つ JWT（HS256）を発行する
func issueToken(userID int64) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
