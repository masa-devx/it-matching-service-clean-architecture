package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/domain"
)

// apiFailure は掲載状態遷移3エンドポイント共通のエラー表現
type apiFailure struct {
	code int
	msg  string
}

// changeProjectStatus は publish / unpublish / close の共通処理。
// 3エンドポイントの違いは「遷移先の状態」だけで、可否の判断は遷移表（domain）が持つ
func (h *Handler) changeProjectStatus(ctx context.Context, projectID int64, to domain.ProjectStatus) (company.TsunaguWorksProject, *apiFailure) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return company.TsunaguWorksProject{}, &apiFailure{code: http.StatusUnauthorized, msg: "認証が必要です"}
	}

	project, err := h.project.ChangeStatus(ctx, claims.UserID, projectID, to)
	if err != nil {
		var transitionErr *usecase.TransitionError
		switch {
		case errors.Is(err, usecase.ErrAuthFailed):
			return company.TsunaguWorksProject{}, &apiFailure{code: http.StatusUnauthorized, msg: "認証が必要です"}
		case errors.Is(err, usecase.ErrProjectNotFound):
			return company.TsunaguWorksProject{}, &apiFailure{code: http.StatusNotFound, msg: err.Error()}
		case errors.As(err, &transitionErr):
			return company.TsunaguWorksProject{}, &apiFailure{code: http.StatusConflict, msg: transitionErr.Error()}
		case errors.Is(err, usecase.ErrStatusConflict):
			return company.TsunaguWorksProject{}, &apiFailure{code: http.StatusConflict, msg: err.Error()}
		default:
			log.Printf("changeProjectStatus: %v", err)
			return company.TsunaguWorksProject{}, &apiFailure{code: http.StatusInternalServerError, msg: "状態の変更に失敗しました"}
		}
	}

	return toAPIProject(project), nil
}

func (h *Handler) ProjectsPublish(ctx context.Context, req company.ProjectsPublishRequestObject) (company.ProjectsPublishResponseObject, error) {
	project, fail := h.changeProjectStatus(ctx, req.Id, domain.ProjectPublished)
	if fail != nil {
		return company.ProjectsPublishdefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: fail.msg},
			StatusCode: fail.code,
		}, nil
	}
	return company.ProjectsPublish200JSONResponse(project), nil
}

func (h *Handler) ProjectsUnpublish(ctx context.Context, req company.ProjectsUnpublishRequestObject) (company.ProjectsUnpublishResponseObject, error) {
	project, fail := h.changeProjectStatus(ctx, req.Id, domain.ProjectDraft)
	if fail != nil {
		return company.ProjectsUnpublishdefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: fail.msg},
			StatusCode: fail.code,
		}, nil
	}
	return company.ProjectsUnpublish200JSONResponse(project), nil
}

func (h *Handler) ProjectsClose(ctx context.Context, req company.ProjectsCloseRequestObject) (company.ProjectsCloseResponseObject, error) {
	project, fail := h.changeProjectStatus(ctx, req.Id, domain.ProjectClosed)
	if fail != nil {
		return company.ProjectsClosedefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: fail.msg},
			StatusCode: fail.code,
		}, nil
	}
	return company.ProjectsClose200JSONResponse(project), nil
}
