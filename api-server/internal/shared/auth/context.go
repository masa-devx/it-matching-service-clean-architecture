package auth

import "context"

// ctxKey は非公開型のキー。他パッケージから同じキーを作れないため、
// context 値の衝突・上書き事故がコンパイルレベルで起きない（Go の定石）
type ctxKey struct{}

// WithClaims は検証済みの本人情報を context に載せる（ミドルウェアだけが呼ぶ想定）
func WithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, ctxKey{}, claims)
}

// ClaimsFrom は検証済みトークン由来の本人情報を取り出す。
// user_id / role は必ずここから取得する（リクエスト値を信用しない＝IDOR対策の入口一本化）
func ClaimsFrom(ctx context.Context) (Claims, bool) {
	claims, ok := ctx.Value(ctxKey{}).(Claims)
	return claims, ok
}
