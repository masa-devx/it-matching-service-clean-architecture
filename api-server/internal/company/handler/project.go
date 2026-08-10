// Package handler は company API（生成された StrictServerInterface）の実装を置く。
package handler

import (
	"context"
	"log"
	"net/http"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
)

// Handler は company API のハンドラ実装。依存（Queries）は main から手渡しされる
type Handler struct {
	queries *db.Queries
}

func New(queries *db.Queries) *Handler {
	return &Handler{queries: queries}
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
	row, err := h.queries.CreateProject(ctx, db.CreateProjectParams{
		Title:          input.Title,
		Description:    input.Description,
		HourlyRateMin:  input.HourlyRateMin,
		HourlyRateMax:  input.HourlyRateMax,
		HoursPerWeek:   input.HoursPerWeek,
		RemoteOk:       input.RemoteOk,
		RequiredSkills: input.RequiredSkills,
	})
	if err != nil {
		// 内部エラーの詳細はログのみに残し、クライアントには安全な文言だけを返す
		log.Printf("CreateProject failed: %v", err)
		return company.ProjectsCreatedefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "案件の作成に失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return company.ProjectsCreate201JSONResponse{
		Id:             row.ID,
		Title:          row.Title,
		Description:    row.Description,
		HourlyRateMin:  row.HourlyRateMin,
		HourlyRateMax:  row.HourlyRateMax,
		HoursPerWeek:   row.HoursPerWeek,
		RemoteOk:       row.RemoteOk,
		RequiredSkills: row.RequiredSkills,
		Status:         company.TsunaguWorksProjectStatus(row.Status),
		CreatedAt:      row.CreatedAt,
	}, nil
}
