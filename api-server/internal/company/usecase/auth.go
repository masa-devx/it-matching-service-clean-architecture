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

// TxBeginner はトランザクションを開始できるもの。
// *pgxpool.Pool（本番）も pgx.Tx（テスト・SAVEPOINT の入れ子になる）も満たす
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// 認証フローの業務エラーは視点共通の語彙として shared/auth に集約し（#30）、再輸出する
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

type SignupCompanyParams struct {
	Email       string
	Password    string
	Name        string
	Location    string
	Description string
}

// SignupCompany は user と company プロフィールを1トランザクションで作成する。
// 途中で失敗したら両方とも残らない（「ユーザーだけ存在してプロフィールが無い」中途半端を作らない）
func (u *Auth) SignupCompany(ctx context.Context, p SignupCompanyParams) (db.User, db.Company, error) {
	hash, err := auth.HashPassword(p.Password)
	if err != nil {
		return db.User{}, db.Company{}, fmt.Errorf("パスワードのハッシュ化に失敗: %w", err)
	}

	tx, err := u.txdb.Begin(ctx)
	if err != nil {
		return db.User{}, db.Company{}, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	// Commit 済みなら Rollback は無害な no-op: エラー経路だけを巻き戻す Go の定石
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := u.queries.WithTx(tx)

	user, err := qtx.CreateUser(ctx, db.CreateUserParams{
		Email:        p.Email,
		PasswordHash: hash,
		Role:         auth.RoleCompany,
	})
	if err != nil {
		if infra.IsUniqueViolation(err) {
			return db.User{}, db.Company{}, ErrEmailTaken
		}
		return db.User{}, db.Company{}, fmt.Errorf("ユーザー作成に失敗: %w", err)
	}

	comp, err := qtx.CreateCompany(ctx, db.CreateCompanyParams{
		UserID:      user.ID,
		Name:        p.Name,
		Location:    p.Location,
		Description: p.Description,
	})
	if err != nil {
		return db.User{}, db.Company{}, fmt.Errorf("企業プロフィール作成に失敗: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return db.User{}, db.Company{}, fmt.Errorf("コミットに失敗: %w", err)
	}
	return user, comp, nil
}

// LoginCompany は照合に成功した企業ユーザーを返す。
// 不存在・パスワード不一致・ロール違いはすべて ErrAuthFailed に潰す（情報を漏らさない）
func (u *Auth) LoginCompany(ctx context.Context, email, password string) (db.User, error) {
	user, err := u.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// 不存在でも bcrypt を1回実行して応答時間を揃える
			auth.VerifyPasswordWithDummy(password)
			return db.User{}, ErrAuthFailed
		}
		return db.User{}, fmt.Errorf("ユーザー取得に失敗: %w", err)
	}

	if err := auth.VerifyPassword(user.PasswordHash, password); err != nil {
		return db.User{}, ErrAuthFailed
	}
	if user.Role != auth.RoleCompany {
		return db.User{}, ErrAuthFailed
	}
	return user, nil
}

// MeCompany はログイン中ユーザーの情報（user + company プロフィール）を返す
func (u *Auth) MeCompany(ctx context.Context, userID int64) (db.User, db.Company, error) {
	user, err := u.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, db.Company{}, ErrAuthFailed
		}
		return db.User{}, db.Company{}, fmt.Errorf("ユーザー取得に失敗: %w", err)
	}

	comp, err := u.queries.GetCompanyByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, db.Company{}, ErrAuthFailed
		}
		return db.User{}, db.Company{}, fmt.Errorf("プロフィール取得に失敗: %w", err)
	}
	return user, comp, nil
}
