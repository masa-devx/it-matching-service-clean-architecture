// Package validator は company API の入力検証を置く。
//
// 検証の役割分担: 単一フィールドの制約（必須・maxLength 等）は仕様（OpenAPI）が一次情報で、
// フォームの生成 Zod と DB の CHECK 制約が既に守っている。
// ここには仕様では表現できない相関ルール（フィールド間の関係）だけを書く。
package validator

import (
	"errors"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
)

// ErrHourlyRateRange は時給の下限が上限を超えているときの検証エラー
var ErrHourlyRateRange = errors.New("hourly_rate_min は hourly_rate_max 以下にしてください")

// HourlyRateRange は時給レンジの相関を検証する（作成・編集で共用）。
// どちらか一方のみ・両方未設定は許可（optional のため）
func HourlyRateRange(minRate, maxRate *int32) error {
	if minRate != nil && maxRate != nil && *minRate > *maxRate {
		return ErrHourlyRateRange
	}
	return nil
}

// CreateProject は案件作成入力の相関ルールを検証する
func CreateProject(input company.TsunaguWorksProjectCreateInput) error {
	return HourlyRateRange(input.HourlyRateMin, input.HourlyRateMax)
}

// UpdateProject は案件編集入力の相関ルールを検証する（作成と同じ規則）
func UpdateProject(input company.TsunaguWorksProjectUpdateInput) error {
	return HourlyRateRange(input.HourlyRateMin, input.HourlyRateMax)
}
