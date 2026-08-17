package handler

import (
	"context"
	"net/http"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
)

// 一覧・詳細・編集のスタブ実装（Step 2 で本実装へ差し替える）

func (h *Handler) ProjectsList(ctx context.Context, req company.ProjectsListRequestObject) (company.ProjectsListResponseObject, error) {
	return company.ProjectsListdefaultJSONResponse{
		Body:       company.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}

func (h *Handler) ProjectsGet(ctx context.Context, req company.ProjectsGetRequestObject) (company.ProjectsGetResponseObject, error) {
	return company.ProjectsGetdefaultJSONResponse{
		Body:       company.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}

func (h *Handler) ProjectsUpdate(ctx context.Context, req company.ProjectsUpdateRequestObject) (company.ProjectsUpdateResponseObject, error) {
	return company.ProjectsUpdatedefaultJSONResponse{
		Body:       company.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}
