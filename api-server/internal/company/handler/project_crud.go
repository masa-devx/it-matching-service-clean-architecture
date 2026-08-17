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

// projectFailure は案件系エンドポイント共通のエラー→HTTP変換（規約の一元化）
func projectFailure(op string, err error) *apiFailure {
	var transitionErr *usecase.TransitionError
	switch {
	case errors.Is(err, usecase.ErrAuthFailed):
		return &apiFailure{code: http.StatusUnauthorized, msg: "認証が必要です"}
	case errors.Is(err, usecase.ErrProjectNotFound):
		return &apiFailure{code: http.StatusNotFound, msg: err.Error()}
	case errors.As(err, &transitionErr):
		return &apiFailure{code: http.StatusConflict, msg: transitionErr.Error()}
	case errors.Is(err, usecase.ErrStatusConflict):
		return &apiFailure{code: http.StatusConflict, msg: err.Error()}
	default:
		log.Printf("%s: %v", op, err)
		return &apiFailure{code: http.StatusInternalServerError, msg: "処理に失敗しました"}
	}
}

func (h *Handler) ProjectsList(ctx context.Context, req company.ProjectsListRequestObject) (company.ProjectsListResponseObject, error) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return company.ProjectsListdefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "認証が必要です"},
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	projects, err := h.project.ListMine(ctx, claims.UserID)
	if err != nil {
		fail := projectFailure("ProjectsList", err)
		return company.ProjectsListdefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: fail.msg},
			StatusCode: fail.code,
		}, nil
	}

	items := make([]company.TsunaguWorksProject, len(projects))
	for i, p := range projects {
		items[i] = toAPIProject(p)
	}
	return company.ProjectsList200JSONResponse{Projects: items}, nil
}

func (h *Handler) ProjectsGet(ctx context.Context, req company.ProjectsGetRequestObject) (company.ProjectsGetResponseObject, error) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return company.ProjectsGetdefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "認証が必要です"},
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	project, err := h.project.GetMine(ctx, claims.UserID, req.Id)
	if err != nil {
		fail := projectFailure("ProjectsGet", err)
		return company.ProjectsGetdefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: fail.msg},
			StatusCode: fail.code,
		}, nil
	}
	return company.ProjectsGet200JSONResponse(toAPIProject(project)), nil
}

func (h *Handler) ProjectsUpdate(ctx context.Context, req company.ProjectsUpdateRequestObject) (company.ProjectsUpdateResponseObject, error) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return company.ProjectsUpdatedefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "認証が必要です"},
			StatusCode: http.StatusUnauthorized,
		}, nil
	}
	if req.Body == nil {
		return company.ProjectsUpdatedefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "リクエストボディが必要です"},
			StatusCode: http.StatusBadRequest,
		}, nil
	}
	input := *req.Body

	if err := validator.UpdateProject(input); err != nil {
		return company.ProjectsUpdatedefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: err.Error()},
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	project, err := h.project.Update(ctx, claims.UserID, req.Id, db.UpdateProjectParams{
		Title:          input.Title,
		Description:    input.Description,
		HourlyRateMin:  input.HourlyRateMin,
		HourlyRateMax:  input.HourlyRateMax,
		HoursPerWeek:   input.HoursPerWeek,
		RemoteOk:       input.RemoteOk,
		RequiredSkills: input.RequiredSkills,
	})
	if err != nil {
		fail := projectFailure("ProjectsUpdate", err)
		return company.ProjectsUpdatedefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: fail.msg},
			StatusCode: fail.code,
		}, nil
	}
	return company.ProjectsUpdate200JSONResponse(toAPIProject(project)), nil
}
