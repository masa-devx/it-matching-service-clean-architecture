package auth_test

import (
	"strings"
	"testing"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
)

// TestPassword は bcrypt によるハッシュ化と照合を固定する。
//
// 目的: 平文パスワードを保存しない・照合が正しく機能する、という認証の前提を守る。
//
// 観点: 照合の成功/失敗・ハッシュに平文が含まれない・
// 同じ平文でも毎回違うハッシュになる（salt が効いている＝レインボーテーブル耐性）。
func TestPassword(t *testing.T) {
	const plain = "correct-horse-battery"

	hash, err := auth.HashPassword(plain)
	if err != nil {
		t.Fatalf("ハッシュ化に失敗: %v", err)
	}

	t.Run("正しいパスワードは照合に成功する", func(t *testing.T) {
		if err := auth.VerifyPassword(hash, plain); err != nil {
			t.Errorf("照合に失敗: %v", err)
		}
	})

	t.Run("誤ったパスワードは照合に失敗する", func(t *testing.T) {
		if err := auth.VerifyPassword(hash, "wrong-password"); err == nil {
			t.Error("誤ったパスワードが照合を通過した")
		}
	})

	t.Run("ハッシュに平文が含まれない", func(t *testing.T) {
		if strings.Contains(hash, plain) {
			t.Error("ハッシュに平文が含まれている")
		}
	})

	t.Run("同じ平文でも毎回違うハッシュになる（salt）", func(t *testing.T) {
		hash2, err := auth.HashPassword(plain)
		if err != nil {
			t.Fatalf("ハッシュ化に失敗: %v", err)
		}
		if hash == hash2 {
			t.Error("同じハッシュが生成された（salt が効いていない）")
		}
	})
}
