package infra

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueViolation は PostgreSQL の UNIQUE 制約違反（SQLSTATE 23505）か判定する。
// 事前 SELECT で重複チェックしない（確認と挿入の間に他リクエストが割り込む TOCTOU を避け、
// 制約違反を「起きてから翻訳する」tsunagu-works の型）
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
