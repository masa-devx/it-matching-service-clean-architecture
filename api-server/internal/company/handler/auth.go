package handler

import (
	"context"
	"net/http"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
)

// 認証系のスタブ実装（Step 2 の usecase 完成後に本実装へ差し替える）。
// 仕様追加の時点でメソッドを生やしておくことで、StrictServerInterface の
// コンパイル検査（project.go の var _ 宣言）を通し、各ステップを単体でビルド可能に保つ

func (h *Handler) AuthSignup(ctx context.Context, req company.AuthSignupRequestObject) (company.AuthSignupResponseObject, error) {
	return company.AuthSignupdefaultJSONResponse{
		Body:       company.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}

func (h *Handler) AuthLogin(ctx context.Context, req company.AuthLoginRequestObject) (company.AuthLoginResponseObject, error) {
	return company.AuthLogindefaultJSONResponse{
		Body:       company.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}

func (h *Handler) AuthMe(ctx context.Context, req company.AuthMeRequestObject) (company.AuthMeResponseObject, error) {
	return company.AuthMedefaultJSONResponse{
		Body:       company.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}
