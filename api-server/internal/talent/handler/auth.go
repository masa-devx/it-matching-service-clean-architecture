// Package handler は talent API（生成された StrictServerInterface）の実装を置く。
//
// エラー変換の規約は company 側と同一: validator → 400 / ErrEmailTaken → 409 /
// ErrAuthFailed → 401 / その他 → 500（詳細はログのみ）
package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	talent "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/talent"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/talent/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/talent/validator"
)

const tokenTTL = 24 * time.Hour

// Handler は talent API のハンドラ実装。依存は main から手渡しされる
type Handler struct {
	auth      *usecase.Auth
	project   *usecase.Project
	jwtSecret []byte
}

func New(authUsecase *usecase.Auth, projectUsecase *usecase.Project, jwtSecret []byte) *Handler {
	return &Handler{auth: authUsecase, project: projectUsecase, jwtSecret: jwtSecret}
}

// 実装漏れをコンパイルエラーにする
var _ talent.StrictServerInterface = (*Handler)(nil)

func (h *Handler) AuthSignup(ctx context.Context, req talent.AuthSignupRequestObject) (talent.AuthSignupResponseObject, error) {
	if req.Body == nil {
		return talent.AuthSignupdefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "リクエストボディが必要です"},
			StatusCode: http.StatusBadRequest,
		}, nil
	}
	input := *req.Body

	if err := validator.Signup(input); err != nil {
		return talent.AuthSignupdefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: err.Error()},
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	user, _, err := h.auth.SignupTalent(ctx, usecase.SignupTalentParams{
		Email:       string(input.Email),
		Password:    input.Password,
		DisplayName: input.DisplayName,
		Skills:      input.Skills,
		Bio:         deref(input.Bio),
	})
	if err != nil {
		if errors.Is(err, usecase.ErrEmailTaken) {
			return talent.AuthSignupdefaultJSONResponse{
				Body:       talent.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusConflict,
			}, nil
		}
		log.Printf("talent AuthSignup: %v", err)
		return talent.AuthSignupdefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "サインアップに失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	token, err := auth.IssueToken(h.jwtSecret, user.ID, user.Role, tokenTTL)
	if err != nil {
		log.Printf("talent AuthSignup: トークン発行に失敗: %v", err)
		return talent.AuthSignupdefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "サインアップに失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return talent.AuthSignup201JSONResponse{Token: token}, nil
}

func (h *Handler) AuthLogin(ctx context.Context, req talent.AuthLoginRequestObject) (talent.AuthLoginResponseObject, error) {
	if req.Body == nil {
		return talent.AuthLogindefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "リクエストボディが必要です"},
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	user, err := h.auth.LoginTalent(ctx, string(req.Body.Email), req.Body.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrAuthFailed) {
			return talent.AuthLogindefaultJSONResponse{
				Body:       talent.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusUnauthorized,
			}, nil
		}
		log.Printf("talent AuthLogin: %v", err)
		return talent.AuthLogindefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "ログインに失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	token, err := auth.IssueToken(h.jwtSecret, user.ID, user.Role, tokenTTL)
	if err != nil {
		log.Printf("talent AuthLogin: トークン発行に失敗: %v", err)
		return talent.AuthLogindefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "ログインに失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return talent.AuthLogin200JSONResponse{Token: token}, nil
}

func (h *Handler) AuthMe(ctx context.Context, req talent.AuthMeRequestObject) (talent.AuthMeResponseObject, error) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return talent.AuthMedefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "認証が必要です"},
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	user, tal, err := h.auth.MeTalent(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, usecase.ErrAuthFailed) {
			return talent.AuthMedefaultJSONResponse{
				Body:       talent.TsunaguWorksApiError{Error: "認証が必要です"},
				StatusCode: http.StatusUnauthorized,
			}, nil
		}
		log.Printf("talent AuthMe: %v", err)
		return talent.AuthMedefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "取得に失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return talent.AuthMe200JSONResponse{
		UserId:      user.ID,
		Email:       user.Email,
		DisplayName: tal.DisplayName,
		Skills:      tal.Skills,
		Bio:         tal.Bio,
	}, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
