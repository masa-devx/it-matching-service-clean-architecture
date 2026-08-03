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

// parseToken はトークン文字列を検証し、sub に入れたユーザーIDを返す。
// 署名不正・期限切れ・形式不正はすべてエラー
func parseToken(tokenString string) (int64, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		// alg すり替え攻撃対策: ヘッダーの alg を鵜呑みにせず、
		// 発行時と同じ HMAC 系であることを検証側で強制する
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("想定外の署名方式です: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return 0, fmt.Errorf("トークン検証失敗: %w", err)
	}
	if !token.Valid {
		return 0, fmt.Errorf("トークンが無効です")
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("sub がユーザーIDとして不正です: %w", err)
	}
	return userID, nil
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
