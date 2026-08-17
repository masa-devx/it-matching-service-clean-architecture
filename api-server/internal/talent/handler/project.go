package handler

import (
	"context"
	"net/http"

	talent "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/talent"
)

// 公開案件の閲覧のスタブ実装（Step 2 で本実装へ差し替える）

func (h *Handler) ProjectsList(ctx context.Context, req talent.ProjectsListRequestObject) (talent.ProjectsListResponseObject, error) {
	return talent.ProjectsListdefaultJSONResponse{
		Body:       talent.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}

func (h *Handler) ProjectsGet(ctx context.Context, req talent.ProjectsGetRequestObject) (talent.ProjectsGetResponseObject, error) {
	return talent.ProjectsGetdefaultJSONResponse{
		Body:       talent.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}
