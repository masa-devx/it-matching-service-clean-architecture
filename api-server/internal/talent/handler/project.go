package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	talent "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/talent"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/talent/usecase"
)

func (h *Handler) ProjectsList(ctx context.Context, req talent.ProjectsListRequestObject) (talent.ProjectsListResponseObject, error) {
	params := usecase.ListPublishedParams{
		Cursor:        req.Params.Cursor,
		Limit:         req.Params.Limit,
		RemoteOk:      req.Params.RemoteOk,
		MinHourlyRate: req.Params.MinHourlyRate,
	}
	if req.Params.Skills != nil {
		params.Skills = *req.Params.Skills
	}

	page, err := h.project.ListPublished(ctx, params)
	if err != nil {
		slog.ErrorContext(ctx, "talent ProjectsList", "err", err)
		return talent.ProjectsListdefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "一覧の取得に失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	items := make([]talent.TsunaguWorksProject, len(page.Projects))
	for i, p := range page.Projects {
		items[i] = toAPIProject(p)
	}
	return talent.ProjectsList200JSONResponse{
		Projects:   items,
		NextCursor: page.NextCursor,
	}, nil
}

func (h *Handler) ProjectsGet(ctx context.Context, req talent.ProjectsGetRequestObject) (talent.ProjectsGetResponseObject, error) {
	project, err := h.project.GetPublished(ctx, req.Id)
	if err != nil {
		if errors.Is(err, usecase.ErrProjectNotFound) {
			return talent.ProjectsGetdefaultJSONResponse{
				Body:       talent.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusNotFound,
			}, nil
		}
		slog.ErrorContext(ctx, "talent ProjectsGet", "err", err)
		return talent.ProjectsGetdefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "取得に失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}
	return talent.ProjectsGet200JSONResponse(toAPIProject(project)), nil
}

// toAPIProject は DB の行を API の型へ詰め替える（talent 視点）
func toAPIProject(p db.Project) talent.TsunaguWorksProject {
	return talent.TsunaguWorksProject{
		Id:             p.ID,
		Title:          p.Title,
		Description:    p.Description,
		HourlyRateMin:  p.HourlyRateMin,
		HourlyRateMax:  p.HourlyRateMax,
		HoursPerWeek:   p.HoursPerWeek,
		RemoteOk:       p.RemoteOk,
		RequiredSkills: p.RequiredSkills,
		Status:         talent.TsunaguWorksProjectStatus(p.Status),
		CreatedAt:      p.CreatedAt,
	}
}
