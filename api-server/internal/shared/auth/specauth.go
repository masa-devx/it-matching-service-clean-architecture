package auth

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// RequiredAuthOps は OpenAPI 仕様から「認証必須の operation」の集合を導出する。
// 一次情報は仕様の security 定義（.tsp の @useAuth）: 認証の要否をコードに手書きしない。
// 起動時に1回だけ呼んで map を作る（毎リクエストで仕様を舐めると線形探索になる・設計プラン§7）
func RequiredAuthOps(spec *openapi3.T) map[string]bool {
	ops := map[string]bool{}
	for _, item := range spec.Paths.Map() {
		for _, op := range item.Operations() {
			if op.Security != nil && len(*op.Security) > 0 {
				ops[goOperationName(op.OperationID)] = true
			}
		}
	}
	return ops
}

// goOperationName は仕様の operationId（例: Auth_me）を、生成コードが
// StrictMiddleware に渡す Go メソッド名（例: AuthMe）へ変換する。
// oapi-codegen の命名規則（アンダースコア区切りを CamelCase 連結）に合わせている
func goOperationName(operationID string) string {
	parts := strings.Split(operationID, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return b.String()
}
