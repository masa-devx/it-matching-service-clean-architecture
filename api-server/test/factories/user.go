package factories

import (
	"fmt"
	"sync/atomic"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
)

// email には UNIQUE 制約があるため、連番で衝突を避ける
// （テスト間は ROLLBACK で分離されるが、同一テスト内で複数ユーザーを作るケースに備える）
var userSeq atomic.Int64

type UserOption func(*db.CreateUserParams)

func WithEmail(email string) UserOption {
	return func(p *db.CreateUserParams) {
		p.Email = email
	}
}

func WithRole(role string) UserOption {
	return func(p *db.CreateUserParams) {
		p.Role = role
	}
}

// CreateUserParams は妥当なデフォルト値のユーザー作成引数を返す。
// password_hash はダミー値（bcrypt の照合を検証するテストは #28 で本物のハッシュを使う）
func CreateUserParams(opts ...UserOption) db.CreateUserParams {
	params := db.CreateUserParams{
		Email:        fmt.Sprintf("user%d@example.com", userSeq.Add(1)),
		PasswordHash: "$2a$10$dummy.hash.for.factories.only",
		Role:         "company",
	}
	for _, opt := range opts {
		opt(&params)
	}
	return params
}
