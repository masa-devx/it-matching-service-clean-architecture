package auth

import (
	"context"
	"net/http"
	"strings"
)

// NewStrictAuth は oapi-codegen の StrictMiddleware 形式の「認証＋ロール認可」を返す。
//
//   - requiredOps: 認証必須の operation 集合。仕様の security 定義から導出する（RequiredAuthOps）。
//     ここに載らない operation は素通り＝認証要否の一次情報はコードではなく仕様
//   - requiredRole: マウント先の視点（/company なら RoleCompany）。パスプレフィックス単位で
//     一律に強制するため、ハンドラにロール判定を書く必要がない（書き忘れが構造的に起きない）
//
// 401 と 403 の区別: 401=誰か分からない（未認証・無効トークン）／403=誰かは分かるが権限が無い。
// 401 の失敗理由（ヘッダー無し・改ざん・期限切れ）は同一レスポンスに潰す（情報を漏らさない）。
//
// 生成される型はパッケージ（company / talent）ごとに別名のため、~func 制約の型パラメータで
// 「同じ形の関数型」を受けて両方に代入できるようにしている
func NewStrictAuth[F ~func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error)](
	secret []byte,
	requiredRole string,
	requiredOps map[string]bool,
) func(next F, operationID string) F {
	return func(next F, operationID string) F {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			if !requiredOps[operationID] {
				return next(ctx, w, r, request)
			}

			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok {
				unauthorized(w)
				return nil, nil // レスポンスは書き込み済み。nil を返すと生成側は何も書かない
			}
			claims, err := VerifyToken(secret, token)
			if err != nil {
				unauthorized(w)
				return nil, nil
			}
			if claims.Role != requiredRole {
				forbidden(w)
				return nil, nil
			}

			return next(WithClaims(ctx, claims), w, r, request)
		}
	}
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"認証が必要です"}`))
}

func forbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte(`{"error":"この操作を行う権限がありません"}`))
}
