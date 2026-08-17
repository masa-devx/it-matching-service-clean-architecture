package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/factories"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/helpers"
)

// TestSignupCompany は企業サインアップの業務フローを実DBで固定する。
//
// 目的: user とプロフィールの原子性（片方だけ残る中途半端を作らない）と、
// email 重複の翻訳（事前 SELECT なし・UNIQUE 違反→ErrEmailTaken）を保証する。
//
// 観点: 正常系（両テーブルに正しく紐づく・パスワードが平文で残らない）／
// 重複 email／**トランザクションの原子性**（プロフィール側の失敗で user も消える
// ＝実DBでしか書けないテスト）。
func TestSignupCompany(t *testing.T) {
	newAuth := func(t *testing.T) (*usecase.Auth, *db.Queries) {
		tx, queries := helpers.NewTestTx(t)
		return usecase.NewAuth(tx, queries), queries
	}

	t.Run("正常系: user と company が作られ、正しく紐づく", func(t *testing.T) {
		uc, _ := newAuth(t)

		user, comp, err := uc.SignupCompany(context.Background(), usecase.SignupCompanyParams{
			Email:    "signup@example.com",
			Password: "password123",
			Name:     "テスト株式会社",
		})
		if err != nil {
			t.Fatalf("サインアップに失敗: %v", err)
		}
		if user.Role != auth.RoleCompany {
			t.Errorf("role: company を期待したが %q", user.Role)
		}
		if comp.UserID != user.ID {
			t.Errorf("company.user_id が user.id と一致しない: %d != %d", comp.UserID, user.ID)
		}
		if strings.Contains(user.PasswordHash, "password123") {
			t.Error("パスワードが平文のまま保存されている")
		}
	})

	t.Run("重複した email は ErrEmailTaken", func(t *testing.T) {
		uc, _ := newAuth(t)

		params := usecase.SignupCompanyParams{Email: "dup@example.com", Password: "password123", Name: "A社"}
		if _, _, err := uc.SignupCompany(context.Background(), params); err != nil {
			t.Fatalf("1回目のサインアップに失敗: %v", err)
		}

		_, _, err := uc.SignupCompany(context.Background(), params)
		if !errors.Is(err, usecase.ErrEmailTaken) {
			t.Errorf("ErrEmailTaken を期待したが: %v", err)
		}
	})

	t.Run("原子性: プロフィール作成が失敗したら user も残らない", func(t *testing.T) {
		uc, queries := newAuth(t)

		// name の DB CHECK（1〜100文字）を意図的に破り、CreateCompany を失敗させる
		_, _, err := uc.SignupCompany(context.Background(), usecase.SignupCompanyParams{
			Email:    "atomic@example.com",
			Password: "password123",
			Name:     strings.Repeat("あ", 101),
		})
		if err == nil {
			t.Fatal("失敗するはずのサインアップが成功した")
		}

		// user だけが残っていない＝トランザクションが巻き戻っている
		_, err = queries.GetUserByEmail(context.Background(), "atomic@example.com")
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Errorf("user が残っている（原子性が壊れている）: %v", err)
		}
	})
}

// TestLoginCompany はログインの成否判定を実DBで固定する。
//
// 目的: 認証の正否と「失敗理由を区別しない」規約を守る。壊れると不正ログインや
// email の存在有無の漏えいにつながる。
//
// 観点: 正常系と、失敗3種（不存在・パスワード不一致・ロール違い）が
// **すべて同一の ErrAuthFailed** になること。
func TestLoginCompany(t *testing.T) {
	setup := func(t *testing.T) *usecase.Auth {
		tx, queries := helpers.NewTestTx(t)
		uc := usecase.NewAuth(tx, queries)

		// company ユーザー（照合可能な本物のハッシュで作る）
		if _, _, err := uc.SignupCompany(context.Background(), usecase.SignupCompanyParams{
			Email: "login@example.com", Password: "password123", Name: "ログイン社",
		}); err != nil {
			t.Fatalf("準備のサインアップに失敗: %v", err)
		}

		// talent ロールのユーザー（ロール違いの検証用）
		hash, err := auth.HashPassword("password123")
		if err != nil {
			t.Fatalf("ハッシュ化に失敗: %v", err)
		}
		params := factories.CreateUserParams(factories.WithEmail("talent@example.com"), factories.WithRole(auth.RoleTalent))
		params.PasswordHash = hash
		if _, err := queries.CreateUser(context.Background(), params); err != nil {
			t.Fatalf("talent ユーザー作成に失敗: %v", err)
		}
		return uc
	}

	t.Run("正しい資格情報でログインできる", func(t *testing.T) {
		uc := setup(t)
		user, err := uc.LoginCompany(context.Background(), "login@example.com", "password123")
		if err != nil {
			t.Fatalf("ログインに失敗: %v", err)
		}
		if user.Email != "login@example.com" {
			t.Errorf("email が一致しない: %q", user.Email)
		}
	})

	fails := []struct {
		name     string
		email    string
		password string
	}{
		{name: "存在しない email", email: "nobody@example.com", password: "password123"},
		{name: "パスワード不一致", email: "login@example.com", password: "wrong-password"},
		{name: "ロール違い（talent が company にログイン）", email: "talent@example.com", password: "password123"},
	}
	for _, tt := range fails {
		t.Run("失敗: "+tt.name, func(t *testing.T) {
			uc := setup(t)
			_, err := uc.LoginCompany(context.Background(), tt.email, tt.password)
			if !errors.Is(err, usecase.ErrAuthFailed) {
				t.Errorf("ErrAuthFailed を期待したが: %v", err)
			}
		})
	}
}

// TestMeCompany はログイン中ユーザー情報の取得を実DBで固定する。
//
// 目的: me が user とプロフィールを正しく合成して返すこと。
// 観点: 正常系・存在しない userID は ErrAuthFailed（トークンは有効だが実体が無いケース）。
func TestMeCompany(t *testing.T) {
	tx, queries := helpers.NewTestTx(t)
	uc := usecase.NewAuth(tx, queries)

	user, comp, err := uc.SignupCompany(context.Background(), usecase.SignupCompanyParams{
		Email: "me@example.com", Password: "password123", Name: "ミー社",
	})
	if err != nil {
		t.Fatalf("準備のサインアップに失敗: %v", err)
	}

	t.Run("正常系: user とプロフィールが取得できる", func(t *testing.T) {
		gotUser, gotComp, err := uc.MeCompany(context.Background(), user.ID)
		if err != nil {
			t.Fatalf("取得に失敗: %v", err)
		}
		if gotUser.ID != user.ID || gotComp.ID != comp.ID {
			t.Errorf("取得結果が一致しない: user=%d company=%d", gotUser.ID, gotComp.ID)
		}
	})

	t.Run("存在しない userID は ErrAuthFailed", func(t *testing.T) {
		_, _, err := uc.MeCompany(context.Background(), 99999999)
		if !errors.Is(err, usecase.ErrAuthFailed) {
			t.Errorf("ErrAuthFailed を期待したが: %v", err)
		}
	})
}
