package factories

import (
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
)

type TalentOption func(*db.CreateTalentParams)

func WithDisplayName(name string) TalentOption {
	return func(p *db.CreateTalentParams) {
		p.DisplayName = name
	}
}

// CreateTalentParams は妥当なデフォルト値の人材プロフィール作成引数を返す。
// user_id は FK のため必須引数にする（company と同じ方針）
func CreateTalentParams(userID int64, opts ...TalentOption) db.CreateTalentParams {
	params := db.CreateTalentParams{
		UserID:      userID,
		DisplayName: "テスト太郎",
		Skills:      []string{"Go", "TypeScript"},
		Bio:         "テスト用の人材です",
	}
	for _, opt := range opts {
		opt(&params)
	}
	return params
}
