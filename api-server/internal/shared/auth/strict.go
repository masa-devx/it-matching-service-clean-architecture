package auth

import (
	"context"
	"net/http"
	"strings"
)

// NewStrictAuth は oapi-codegen の StrictMiddleware 形式の認証ミドルウェアを返す。
// publicOps に載っていない operation はすべて Bearer 検証を必須とし、
// 検証済みの Claims を context に載せて次へ渡す。
//
// 生成される StrictMiddlewareFunc / StrictHandlerFunc はパッケージ（company / talent）ごとに
// 別の名前付き型になるため、型パラメータ F で「同じ形の関数型」を受けて両方に代入できるようにする
func NewStrictAuth[F ~func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error)](
	secret []byte,
	publicOps map[string]bool,
) func(next F, operationID string) F {
	return func(next F, operationID string) F {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			if publicOps[operationID] {
				return next(ctx, w, r, request)
			}

			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok {
				unauthorized(w)
				return nil, nil // 401 は書き込み済み。nil を返すと生成側は何も書かない
			}
			claims, err := VerifyToken(secret, token)
			if err != nil {
				unauthorized(w)
				return nil, nil
			}

			return next(WithClaims(ctx, claims), w, r, request)
		}
	}
}
