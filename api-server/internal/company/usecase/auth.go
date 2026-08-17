package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
)

// TxBeginner はトランザクションを開始できるもの。
// *pgxpool.Pool（本番）も pgx.Tx（テスト・SAVEPOINT の入れ子になる）も満たす
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

var (
	// ErrEmailTaken は email の重複（handler が 409 に変換する）
	ErrEmailTaken = errors.New("このメールアドレスは既に登録されています")
	// ErrAuthFailed は認証失敗。理由（不存在・パスワード不一致・ロール違い）は区別しない
	ErrAuthFailed = errors.New("メールアドレスまたはパスワードが正しくありません")
)

// ユーザー不存在時にも bcrypt を1回実行し、応答時間の差で email の存在有無を漏らさないための
// ダミーハッシュ（同一401の3点セット: 文言・ステータス・応答時間）
var dummyHash string

func init() {
	h, err := auth.HashPassword("timing-equalizer-dummy")
	if err != nil {
		panic(fmt.Sprintf("ダミーハッシュの生成に失敗: %v", err))
	}
	dummyHash = h
}

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
		if isUniqueViolation(err) {
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
			_ = auth.VerifyPassword(dummyHash, password)
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

// isUniqueViolation は PostgreSQL の UNIQUE 制約違反（SQLSTATE 23505）か判定する。
// 事前 SELECT で重複チェックしない（確認と挿入の間に他リクエストが割り込む TOCTOU を避け、
// 制約違反を「起きてから翻訳する」tsunagu-works の型）
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
