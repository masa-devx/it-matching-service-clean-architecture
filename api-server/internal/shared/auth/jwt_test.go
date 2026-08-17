package auth_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
)

var testSecret = []byte("test-secret-for-jwt")

// TestIssueAndVerifyToken は JWT の発行と検証の往復を固定する。
//
// 目的: 認証の土台。壊れると全ての保護エンドポイントが不正なトークンを受け入れる/
// 正当なユーザーを拒否する。
//
// 観点: 正常往復（userID / role の一致）と、拒否すべきトークンの網羅
// （改ざん・期限切れ・別鍵・alg=none・正しく署名された別方式）。
// 特に後半2つは「署名の正しさ」ではなく「方式の決め打ち」を検証する攻撃系ケース。
func TestIssueAndVerifyToken(t *testing.T) {
	t.Run("正常系: 発行したトークンから userID と role が取り出せる", func(t *testing.T) {
		token, err := auth.IssueToken(testSecret, 42, auth.RoleCompany, time.Hour)
		if err != nil {
			t.Fatalf("発行に失敗: %v", err)
		}

		claims, err := auth.VerifyToken(testSecret, token)
		if err != nil {
			t.Fatalf("検証に失敗: %v", err)
		}
		if claims.UserID != 42 || claims.Role != auth.RoleCompany {
			t.Errorf("claims が一致しない: %+v", claims)
		}
	})

	// 拒否すべきトークンのテーブル駆動
	tests := []struct {
		name  string
		token func(t *testing.T) string
	}{
		{
			name: "署名部分を改ざんしたトークン",
			token: func(t *testing.T) string {
				token, _ := auth.IssueToken(testSecret, 42, auth.RoleCompany, time.Hour)
				return token[:len(token)-4] + "xxxx"
			},
		},
		{
			name: "期限切れのトークン",
			token: func(t *testing.T) string {
				token, _ := auth.IssueToken(testSecret, 42, auth.RoleCompany, -time.Hour)
				return token
			},
		},
		{
			name: "別の秘密鍵で署名されたトークン",
			token: func(t *testing.T) string {
				token, _ := auth.IssueToken([]byte("attacker-secret"), 42, auth.RoleCompany, time.Hour)
				return token
			},
		},
		{
			name: "alg=none のトークン（署名検証の無効化を狙う攻撃）",
			token: func(t *testing.T) string {
				claims := jwt.RegisteredClaims{
					Subject:   "42",
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				}
				token, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).
					SignedString(jwt.UnsafeAllowNoneSignatureType)
				if err != nil {
					t.Fatalf("攻撃用トークンの作成に失敗: %v", err)
				}
				return token
			},
		},
		{
			name: "HS512 で正しく署名されたトークン（方式すり替え）",
			token: func(t *testing.T) string {
				claims := jwt.RegisteredClaims{
					Subject:   "42",
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				}
				token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(testSecret)
				if err != nil {
					t.Fatalf("トークンの作成に失敗: %v", err)
				}
				return token
			},
		},
	}

	for _, tt := range tests {
		t.Run("拒否: "+tt.name, func(t *testing.T) {
			_, err := auth.VerifyToken(testSecret, tt.token(t))
			if err == nil {
				t.Fatal("拒否されるべきトークンが受け入れられた")
			}
			if !errors.Is(err, auth.ErrInvalidToken) {
				t.Errorf("ErrInvalidToken を期待したが: %v", err)
			}
		})
	}
}

// TestTokenDoesNotContainSecret はトークン文字列に秘密鍵が漏れないことを固定する。
//
// 目的: JWT は「署名付きだが暗号化はされていない」。誤って秘密情報を claims に
// 入れる事故への感度を保つ（トークンは誰でもデコードできる前提で扱う）。
func TestTokenDoesNotContainSecret(t *testing.T) {
	token, err := auth.IssueToken(testSecret, 42, auth.RoleCompany, time.Hour)
	if err != nil {
		t.Fatalf("発行に失敗: %v", err)
	}
	if strings.Contains(token, string(testSecret)) {
		t.Error("トークンに秘密鍵が含まれている")
	}
}
