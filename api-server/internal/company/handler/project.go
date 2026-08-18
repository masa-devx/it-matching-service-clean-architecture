// Package handler は company API（生成された StrictServerInterface）の実装を置く。
//
// エラー変換の規約: validator のエラー → 400（メッセージは利用者向けなのでそのまま返す）／
// usecase のエラー → 500（詳細はログのみ・クライアントには安全な文言だけ）
package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/validator"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
)

// Handler は company API のハンドラ実装。依存（usecase・JWT秘密鍵）は main から手渡しされる
type Handler struct {
	project     *usecase.Project
	auth        *usecase.Auth
	application *usecase.Application
	jwtSecret   []byte
}

func New(project *usecase.Project, authUsecase *usecase.Auth, applicationUsecase *usecase.Application, jwtSecret []byte) *Handler {
	return &Handler{
		project:     project,
		auth:        authUsecase,
		application: applicationUsecase,
		jwtSecret:   jwtSecret,
	}
}

// 実装漏れをコンパイルエラーにする（仕様にエンドポイントが増えると、ここで検出される）
var _ company.StrictServerInterface = (*Handler)(nil)

// ProjectsCreate は案件を draft として作成する。
// この層の仕事は「検証の呼び出し → 生成型と DB 型の詰め替え → エラーの HTTP 変換」だけ
func (h *Handler) ProjectsCreate(ctx context.Context, req company.ProjectsCreateRequestObject) (company.ProjectsCreateResponseObject, error) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return company.ProjectsCreatedefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "認証が必要です"},
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	if req.Body == nil {
		return company.ProjectsCreatedefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "リクエストボディが必要です"},
			StatusCode: http.StatusBadRequest,
		}, nil
	}
	input := *req.Body

	if err := validator.CreateProject(input); err != nil {
		return company.ProjectsCreatedefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: err.Error()},
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	project, err := h.project.Create(ctx, claims.UserID, db.CreateProjectParams{
		Title:          input.Title,
		Description:    input.Description,
		HourlyRateMin:  input.HourlyRateMin,
		HourlyRateMax:  input.HourlyRateMax,
		HoursPerWeek:   input.HoursPerWeek,
		RemoteOk:       input.RemoteOk,
		RequiredSkills: input.RequiredSkills,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrAuthFailed) {
			return company.ProjectsCreatedefaultJSONResponse{
				Body:       company.TsunaguWorksApiError{Error: "認証が必要です"},
				StatusCode: http.StatusUnauthorized,
			}, nil
		}
		log.Printf("ProjectsCreate: %v", err)
		return company.ProjectsCreatedefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "案件の作成に失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return company.ProjectsCreate201JSONResponse(toAPIProject(project)), nil
}

// toAPIProject は DB の行を API の型へ詰め替える（作成・状態遷移で共用）
func toAPIProject(p db.Project) company.TsunaguWorksProject {
	return company.TsunaguWorksProject{
		Id:             p.ID,
		Title:          p.Title,
		Description:    p.Description,
		HourlyRateMin:  p.HourlyRateMin,
		HourlyRateMax:  p.HourlyRateMax,
		HoursPerWeek:   p.HoursPerWeek,
		RemoteOk:       p.RemoteOk,
		RequiredSkills: p.RequiredSkills,
		Status:         company.TsunaguWorksProjectStatus(p.Status),
		CreatedAt:      p.CreatedAt,
	}
}
