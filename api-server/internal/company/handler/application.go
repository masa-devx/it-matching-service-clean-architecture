package handler

import (
	"context"
	"net/http"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
)

// 選考 API のスタブ（#57 Step 2 で本実装に差し替える）

func (h *Handler) ProjectsListApplications(ctx context.Context, req company.ProjectsListApplicationsRequestObject) (company.ProjectsListApplicationsResponseObject, error) {
	return company.ProjectsListApplicationsdefaultJSONResponse{
		Body:       company.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}

func (h *Handler) ApplicationsOffer(ctx context.Context, req company.ApplicationsOfferRequestObject) (company.ApplicationsOfferResponseObject, error) {
	return company.ApplicationsOfferdefaultJSONResponse{
		Body:       company.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}

func (h *Handler) ApplicationsReject(ctx context.Context, req company.ApplicationsRejectRequestObject) (company.ApplicationsRejectResponseObject, error) {
	return company.ApplicationsRejectdefaultJSONResponse{
		Body:       company.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}
