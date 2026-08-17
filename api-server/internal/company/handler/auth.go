package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/validator"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
)

// tokenTTL はトークンの有効期限。短すぎると再ログインが頻発し、長すぎると漏えい時の被害が広がる
const tokenTTL = 24 * time.Hour

func (h *Handler) AuthSignup(ctx context.Context, req company.AuthSignupRequestObject) (company.AuthSignupResponseObject, error) {
	if req.Body == nil {
		return company.AuthSignupdefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "リクエストボディが必要です"},
			StatusCode: http.StatusBadRequest,
		}, nil
	}
	input := *req.Body

	if err := validator.Signup(input); err != nil {
		return company.AuthSignupdefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: err.Error()},
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	user, _, err := h.auth.SignupCompany(ctx, usecase.SignupCompanyParams{
		Email:       string(input.Email),
		Password:    input.Password,
		Name:        input.Name,
		Location:    deref(input.Location),
		Description: deref(input.Description),
	})
	if err != nil {
		if errors.Is(err, usecase.ErrEmailTaken) {
			return company.AuthSignupdefaultJSONResponse{
				Body:       company.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusConflict,
			}, nil
		}
		log.Printf("AuthSignup: %v", err)
		return company.AuthSignupdefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "サインアップに失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	token, err := auth.IssueToken(h.jwtSecret, user.ID, user.Role, tokenTTL)
	if err != nil {
		log.Printf("AuthSignup: トークン発行に失敗: %v", err)
		return company.AuthSignupdefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "サインアップに失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return company.AuthSignup201JSONResponse{Token: token}, nil
}

func (h *Handler) AuthLogin(ctx context.Context, req company.AuthLoginRequestObject) (company.AuthLoginResponseObject, error) {
	if req.Body == nil {
		return company.AuthLogindefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "リクエストボディが必要です"},
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	user, err := h.auth.LoginCompany(ctx, string(req.Body.Email), req.Body.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrAuthFailed) {
			return company.AuthLogindefaultJSONResponse{
				Body:       company.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusUnauthorized,
			}, nil
		}
		log.Printf("AuthLogin: %v", err)
		return company.AuthLogindefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "ログインに失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	token, err := auth.IssueToken(h.jwtSecret, user.ID, user.Role, tokenTTL)
	if err != nil {
		log.Printf("AuthLogin: トークン発行に失敗: %v", err)
		return company.AuthLogindefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "ログインに失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return company.AuthLogin200JSONResponse{Token: token}, nil
}

func (h *Handler) AuthMe(ctx context.Context, req company.AuthMeRequestObject) (company.AuthMeResponseObject, error) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		// ミドルウェア未適用の経路は存在しない想定だが、安全側に倒す
		return company.AuthMedefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "認証が必要です"},
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	user, comp, err := h.auth.MeCompany(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, usecase.ErrAuthFailed) {
			return company.AuthMedefaultJSONResponse{
				Body:       company.TsunaguWorksApiError{Error: "認証が必要です"},
				StatusCode: http.StatusUnauthorized,
			}, nil
		}
		log.Printf("AuthMe: %v", err)
		return company.AuthMedefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "取得に失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return company.AuthMe200JSONResponse{
		UserId:      user.ID,
		Email:       user.Email,
		Name:        comp.Name,
		Location:    comp.Location,
		Description: comp.Description,
	}, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
