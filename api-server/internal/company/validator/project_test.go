package validator_test

import (
	"testing"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/validator"
)

// TestCreateProject は案件作成入力の相関検証を固定する。
//
// 目的: 「時給の下限 > 上限」という、仕様（OpenAPI）では表現できない矛盾した入力を
// API 境界で弾く。壊れると矛盾した募集条件の案件が作成できてしまう。
//
// 観点: 境界値を両側から張る（min=max は許可 / min=max+1 は拒否）。
// optional フィールドのため「片側のみ」「両方未設定」が検証対象外であることも固定する。
func TestCreateProject(t *testing.T) {
	i32 := func(v int32) *int32 { return &v }

	tests := []struct {
		name    string
		min     *int32
		max     *int32
		wantErr bool
	}{
		{name: "両方未設定は通る", min: nil, max: nil, wantErr: false},
		{name: "下限のみは通る", min: i32(3000), max: nil, wantErr: false},
		{name: "上限のみは通る", min: nil, max: i32(5000), wantErr: false},
		{name: "下限=上限は通る（境界）", min: i32(4000), max: i32(4000), wantErr: false},
		{name: "下限が上限+1なら拒否（境界）", min: i32(4001), max: i32(4000), wantErr: true},
		{name: "下限が上限を大きく超えると拒否", min: i32(9000), max: i32(3000), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := company.TsunaguWorksProjectCreateInput{
				Title:          "テスト案件",
				Description:    "詳細",
				HoursPerWeek:   10,
				RemoteOk:       true,
				RequiredSkills: []string{},
				HourlyRateMin:  tt.min,
				HourlyRateMax:  tt.max,
			}

			err := validator.CreateProject(input)

			if tt.wantErr && err == nil {
				t.Error("エラーになるべき入力が許可された")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("許可されるべき入力が拒否された: %v", err)
			}
		})
	}
}
