// Package usecase は talent API の業務ロジックを置く。
// company 側と対称の構成（Signup / Login / Me の本体は視点ごとに持つ判断・#30）
package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/infra"
)

// TxBeginner はトランザクションを開始できるもの（インターフェースは利用側が定義する Go の慣習）
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// 認証フローの業務エラー（shared/auth の語彙を再輸出）
var (
	ErrEmailTaken = auth.ErrEmailTaken
	ErrAuthFailed = auth.ErrAuthFailed
)

// Auth は認証まわりの業務ロジック
type Auth struct {
	txdb    TxBeginner
	queries *db.Queries
}

func NewAuth(txdb TxBeginner, queries *db.Queries) *Auth {
	return &Auth{txdb: txdb, queries: queries}
}

type SignupTalentParams struct {
	Email       string
	Password    string
	DisplayName string
	Skills      []string
	Bio         string
}

// SignupTalent は user と talent プロフィールを1トランザクションで作成する
func (u *Auth) SignupTalent(ctx context.Context, p SignupTalentParams) (db.User, db.Talent, error) {
	hash, err := auth.HashPassword(p.Password)
	if err != nil {
		return db.User{}, db.Talent{}, fmt.Errorf("パスワードのハッシュ化に失敗: %w", err)
	}

	tx, err := u.txdb.Begin(ctx)
	if err != nil {
		return db.User{}, db.Talent{}, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := u.queries.WithTx(tx)

	// nil スライスは SQL の NULL になり NOT NULL 制約に弾かれる（DEFAULT は列省略時のみ有効）。
	// 「スキル未入力」は空配列として保存する
	skills := p.Skills
	if skills == nil {
		skills = []string{}
	}

	user, err := qtx.CreateUser(ctx, db.CreateUserParams{
		Email:        p.Email,
		PasswordHash: hash,
		Role:         auth.RoleTalent,
	})
	if err != nil {
		if infra.IsUniqueViolation(err) {
			return db.User{}, db.Talent{}, ErrEmailTaken
		}
		return db.User{}, db.Talent{}, fmt.Errorf("ユーザー作成に失敗: %w", err)
	}

	tal, err := qtx.CreateTalent(ctx, db.CreateTalentParams{
		UserID:      user.ID,
		DisplayName: p.DisplayName,
		Skills:      skills,
		Bio:         p.Bio,
	})
	if err != nil {
		return db.User{}, db.Talent{}, fmt.Errorf("人材プロフィール作成に失敗: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.User{}, db.Talent{}, fmt.Errorf("コミットに失敗: %w", err)
	}
	return user, tal, nil
}

// LoginTalent は照合に成功した人材ユーザーを返す。
// 不存在・パスワード不一致・ロール違いはすべて ErrAuthFailed に潰す
func (u *Auth) LoginTalent(ctx context.Context, email, password string) (db.User, error) {
	user, err := u.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			auth.VerifyPasswordWithDummy(password)
			return db.User{}, ErrAuthFailed
		}
		return db.User{}, fmt.Errorf("ユーザー取得に失敗: %w", err)
	}

	if err := auth.VerifyPassword(user.PasswordHash, password); err != nil {
		return db.User{}, ErrAuthFailed
	}
	if user.Role != auth.RoleTalent {
		return db.User{}, ErrAuthFailed
	}
	return user, nil
}

// MeTalent はログイン中ユーザーの情報（user + talent プロフィール）を返す
func (u *Auth) MeTalent(ctx context.Context, userID int64) (db.User, db.Talent, error) {
	user, err := u.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, db.Talent{}, ErrAuthFailed
		}
		return db.User{}, db.Talent{}, fmt.Errorf("ユーザー取得に失敗: %w", err)
	}

	tal, err := u.queries.GetTalentByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, db.Talent{}, ErrAuthFailed
		}
		return db.User{}, db.Talent{}, fmt.Errorf("プロフィール取得に失敗: %w", err)
	}
	return user, tal, nil
}
