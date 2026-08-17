package handler

import (
	"context"
	"net/http"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
)

// 掲載状態遷移のスタブ実装（Step 3 の usecase 完成後に本実装へ差し替える）

func (h *Handler) ProjectsPublish(ctx context.Context, req company.ProjectsPublishRequestObject) (company.ProjectsPublishResponseObject, error) {
	return company.ProjectsPublishdefaultJSONResponse{
		Body:       company.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}

func (h *Handler) ProjectsUnpublish(ctx context.Context, req company.ProjectsUnpublishRequestObject) (company.ProjectsUnpublishResponseObject, error) {
	return company.ProjectsUnpublishdefaultJSONResponse{
		Body:       company.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}

func (h *Handler) ProjectsClose(ctx context.Context, req company.ProjectsCloseRequestObject) (company.ProjectsCloseResponseObject, error) {
	return company.ProjectsClosedefaultJSONResponse{
		Body:       company.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}
