// Package auth は認証・認可の共通部品（JWT・パスワード・context・ミドルウェア）を置く。
package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ユーザーの種別（users.role と同じ語彙・ユビキタス言語）
const (
	RoleCompany = "company"
	RoleTalent  = "talent"
)

// ErrInvalidToken はトークン検証の失敗。原因（改ざん・期限切れ等）は呼び出し側に区別させない
var ErrInvalidToken = errors.New("トークンが無効です")

// Claims は検証済みトークンから取り出せる本人情報
type Claims struct {
	UserID int64
	Role   string
}

type jwtClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// IssueToken は userID と role を含む JWT（HS256）を発行する
func IssueToken(secret []byte, userID int64, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

// VerifyToken はトークンを検証して Claims を返す。
// 署名方式は HS256 に固定する: トークン側が名乗る alg を信用すると
// 署名検証を無効化できてしまう（alg=none / 方式すり替え）ため、検証側が決め打ちする
func VerifyToken(secret []byte, tokenString string) (Claims, error) {
	var claims jwtClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims,
		func(t *jwt.Token) (any, error) { return secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		// 多重ラップ: 呼び出し側の分岐は ErrInvalidToken に一本化しつつ、原因（期限切れ等）もログで追える
		return Claims{}, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return Claims{}, fmt.Errorf("%w: subject が数値ではない", ErrInvalidToken)
	}
	if claims.Role != RoleCompany && claims.Role != RoleTalent {
		return Claims{}, fmt.Errorf("%w: 不明な role", ErrInvalidToken)
	}

	return Claims{UserID: userID, Role: claims.Role}, nil
}
