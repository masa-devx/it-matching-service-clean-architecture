// Package factories はテストデータの生成を置く。
// 「妥当なデフォルト値」を1か所で定義し、テストは関心のある項目だけを上書きする。
// デフォルトが変わってもテストが壊れない（テストの意図が上書き項目として明示される）
package factories

import (
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
)

// ProjectOption はデフォルト値への上書き
type ProjectOption func(*db.CreateProjectParams)

func WithHourlyRate(minRate, maxRate int32) ProjectOption {
	return func(p *db.CreateProjectParams) {
		p.HourlyRateMin = &minRate
		p.HourlyRateMax = &maxRate
	}
}

func WithHoursPerWeek(hours int32) ProjectOption {
	return func(p *db.CreateProjectParams) {
		p.HoursPerWeek = hours
	}
}

// CreateProjectParams は妥当なデフォルト値の案件作成引数を返す
func CreateProjectParams(opts ...ProjectOption) db.CreateProjectParams {
	params := db.CreateProjectParams{
		Title:          "テスト案件",
		Description:    "テスト用の説明",
		HoursPerWeek:   10,
		RemoteOk:       true,
		RequiredSkills: []string{"Go", "PostgreSQL"},
	}
	for _, opt := range opts {
		opt(&params)
	}
	return params
}
