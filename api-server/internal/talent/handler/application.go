package handler

import (
	"context"
	"net/http"

	talent "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/talent"
)

// 応募 API のスタブ（#56 Step 2 で本実装に差し替える）。
// 生成 IF を満たすことでコンパイルを通し、仕様との差分を段階的に埋める

func (h *Handler) ApplicationsCreate(ctx context.Context, req talent.ApplicationsCreateRequestObject) (talent.ApplicationsCreateResponseObject, error) {
	return talent.ApplicationsCreatedefaultJSONResponse{
		Body:       talent.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}

func (h *Handler) ApplicationsList(ctx context.Context, req talent.ApplicationsListRequestObject) (talent.ApplicationsListResponseObject, error) {
	return talent.ApplicationsListdefaultJSONResponse{
		Body:       talent.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}

func (h *Handler) ApplicationsWithdraw(ctx context.Context, req talent.ApplicationsWithdrawRequestObject) (talent.ApplicationsWithdrawResponseObject, error) {
	return talent.ApplicationsWithdrawdefaultJSONResponse{
		Body:       talent.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}
