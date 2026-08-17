package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/talent/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/factories"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/test/helpers"
)

// TestSignupTalent は人材サインアップの業務フローを実DBで固定する。
//
// 目的: company 側（#29）と対称の保証——user と talent プロフィールの原子性・
// email 重複の翻訳——を talent 系統でも守る。
//
// 観点: 正常系（role=talent・skills の TEXT[] 往復）／重複 email／
// 原子性（display_name の CHECK[1〜50] を51文字で破り、user も残らないこと）。
func TestSignupTalent(t *testing.T) {
	t.Run("正常系: user と talent が作られ、正しく紐づく", func(t *testing.T) {
		tx, queries := helpers.NewTestTx(t)
		uc := usecase.NewAuth(tx, queries)

		user, tal, err := uc.SignupTalent(context.Background(), usecase.SignupTalentParams{
			Email:       "talent-signup@example.com",
			Password:    "password123",
			DisplayName: "山田太郎",
			Skills:      []string{"Go", "TypeScript"},
		})
		if err != nil {
			t.Fatalf("サインアップに失敗: %v", err)
		}
		if user.Role != auth.RoleTalent {
			t.Errorf("role: talent を期待したが %q", user.Role)
		}
		if tal.UserID != user.ID {
			t.Errorf("talent.user_id が user.id と一致しない")
		}
		if len(tal.Skills) != 2 || tal.Skills[0] != "Go" {
			t.Errorf("skills が往復していない: %v", tal.Skills)
		}
	})

	t.Run("重複した email は ErrEmailTaken", func(t *testing.T) {
		tx, queries := helpers.NewTestTx(t)
		uc := usecase.NewAuth(tx, queries)

		params := usecase.SignupTalentParams{Email: "talent-dup@example.com", Password: "password123", DisplayName: "重複"}
		if _, _, err := uc.SignupTalent(context.Background(), params); err != nil {
			t.Fatalf("1回目のサインアップに失敗: %v", err)
		}
		_, _, err := uc.SignupTalent(context.Background(), params)
		if !errors.Is(err, usecase.ErrEmailTaken) {
			t.Errorf("ErrEmailTaken を期待したが: %v", err)
		}
	})

	t.Run("原子性: プロフィール作成が失敗したら user も残らない", func(t *testing.T) {
		tx, queries := helpers.NewTestTx(t)
		uc := usecase.NewAuth(tx, queries)

		_, _, err := uc.SignupTalent(context.Background(), usecase.SignupTalentParams{
			Email:       "talent-atomic@example.com",
			Password:    "password123",
			DisplayName: strings.Repeat("あ", 51), // CHECK は 1〜50 文字
		})
		if err == nil {
			t.Fatal("失敗するはずのサインアップが成功した")
		}

		_, err = queries.GetUserByEmail(context.Background(), "talent-atomic@example.com")
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Errorf("user が残っている（原子性が壊れている）: %v", err)
		}
	})
}

// TestLoginTalent はログインの成否判定を実DBで固定する。
//
// 目的: 認証の正否と「失敗理由を区別しない」規約（company 側と同一）。
// 観点: 正常系と、失敗3種（不存在・パスワード不一致・**company ロールの拒否**）が
// すべて同一の ErrAuthFailed になること。
func TestLoginTalent(t *testing.T) {
	setup := func(t *testing.T) *usecase.Auth {
		tx, queries := helpers.NewTestTx(t)
		uc := usecase.NewAuth(tx, queries)

		if _, _, err := uc.SignupTalent(context.Background(), usecase.SignupTalentParams{
			Email: "talent-login@example.com", Password: "password123", DisplayName: "ログイン",
		}); err != nil {
			t.Fatalf("準備のサインアップに失敗: %v", err)
		}

		// company ロールのユーザー（ロール違いの検証用）
		hash, err := auth.HashPassword("password123")
		if err != nil {
			t.Fatalf("ハッシュ化に失敗: %v", err)
		}
		params := factories.CreateUserParams(factories.WithEmail("company-user@example.com"), factories.WithRole(auth.RoleCompany))
		params.PasswordHash = hash
		if _, err := queries.CreateUser(context.Background(), params); err != nil {
			t.Fatalf("company ユーザー作成に失敗: %v", err)
		}
		return uc
	}

	t.Run("正しい資格情報でログインできる", func(t *testing.T) {
		uc := setup(t)
		user, err := uc.LoginTalent(context.Background(), "talent-login@example.com", "password123")
		if err != nil {
			t.Fatalf("ログインに失敗: %v", err)
		}
		if user.Role != auth.RoleTalent {
			t.Errorf("role が talent でない: %q", user.Role)
		}
	})

	fails := []struct {
		name     string
		email    string
		password string
	}{
		{name: "存在しない email", email: "nobody@example.com", password: "password123"},
		{name: "パスワード不一致", email: "talent-login@example.com", password: "wrong-password"},
		{name: "ロール違い（company が talent にログイン）", email: "company-user@example.com", password: "password123"},
	}
	for _, tt := range fails {
		t.Run("失敗: "+tt.name, func(t *testing.T) {
			uc := setup(t)
			_, err := uc.LoginTalent(context.Background(), tt.email, tt.password)
			if !errors.Is(err, usecase.ErrAuthFailed) {
				t.Errorf("ErrAuthFailed を期待したが: %v", err)
			}
		})
	}
}

// TestMeTalent はログイン中ユーザー情報の取得を実DBで固定する。
func TestMeTalent(t *testing.T) {
	tx, queries := helpers.NewTestTx(t)
	uc := usecase.NewAuth(tx, queries)

	user, tal, err := uc.SignupTalent(context.Background(), usecase.SignupTalentParams{
		Email: "talent-me@example.com", Password: "password123", DisplayName: "ミー", Skills: []string{"Go"},
	})
	if err != nil {
		t.Fatalf("準備のサインアップに失敗: %v", err)
	}

	gotUser, gotTal, err := uc.MeTalent(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("取得に失敗: %v", err)
	}
	if gotUser.ID != user.ID || gotTal.ID != tal.ID {
		t.Errorf("取得結果が一致しない")
	}
}
