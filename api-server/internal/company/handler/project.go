// Package handler は company API（生成された StrictServerInterface）の実装を置く。
package handler

import (
	"context"
	"net/http"
	"time"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
)

// Handler は company API のハンドラ実装。Phase 0 は仮実装（永続化しない・DB接続は #7）
type Handler struct{}

func New() *Handler {
	return &Handler{}
}

// 実装漏れをコンパイルエラーにする（仕様にエンドポイントが増えると、ここで検出される）
var _ company.StrictServerInterface = (*Handler)(nil)

// ProjectsCreate は案件を draft として作成する。
func (h *Handler) ProjectsCreate(ctx context.Context, req company.ProjectsCreateRequestObject) (company.ProjectsCreateResponseObject, error) {
	if req.Body == nil {
		return company.ProjectsCreatedefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "リクエストボディが必要です"},
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	input := req.Body
	return company.ProjectsCreate201JSONResponse{
		Id:             1, // 仮実装の固定値。#7 で DB の採番に置き換える
		Title:          input.Title,
		Description:    input.Description,
		HourlyRateMin:  input.HourlyRateMin,
		HourlyRateMax:  input.HourlyRateMax,
		HoursPerWeek:   input.HoursPerWeek,
		RemoteOk:       input.RemoteOk,
		RequiredSkills: input.RequiredSkills,
		Status:         company.Draft,
		CreatedAt:      time.Now().UTC(),
	}, nil
}
