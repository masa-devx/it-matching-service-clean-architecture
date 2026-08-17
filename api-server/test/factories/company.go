package factories

import (
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
)

type CompanyOption func(*db.CreateCompanyParams)

func WithCompanyName(name string) CompanyOption {
	return func(p *db.CreateCompanyParams) {
		p.Name = name
	}
}

// CreateCompanyParams は妥当なデフォルト値の企業プロフィール作成引数を返す。
// user_id は FK のため必須引数にする（factories 側で勝手にユーザーを作らない＝Txの管理はテストが持つ）
func CreateCompanyParams(userID int64, opts ...CompanyOption) db.CreateCompanyParams {
	params := db.CreateCompanyParams{
		UserID:      userID,
		Name:        "テスト株式会社",
		Location:    "東京",
		Description: "テスト用の企業です",
	}
	for _, opt := range opts {
		opt(&params)
	}
	return params
}
