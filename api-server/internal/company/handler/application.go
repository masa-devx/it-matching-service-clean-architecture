package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	company "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/company"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/company/usecase"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/auth"
)

func (h *Handler) ProjectsListApplications(ctx context.Context, req company.ProjectsListApplicationsRequestObject) (company.ProjectsListApplicationsResponseObject, error) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return company.ProjectsListApplicationsdefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "認証が必要です"},
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	rows, err := h.application.ListForProject(ctx, claims.UserID, req.Id)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrProjectNotFound):
			return company.ProjectsListApplicationsdefaultJSONResponse{
				Body:       company.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusNotFound,
			}, nil
		case errors.Is(err, usecase.ErrAuthFailed):
			return company.ProjectsListApplicationsdefaultJSONResponse{
				Body:       company.TsunaguWorksApiError{Error: "認証が必要です"},
				StatusCode: http.StatusUnauthorized,
			}, nil
		}
		slog.ErrorContext(ctx, "company ProjectsListApplications", "err", err)
		return company.ProjectsListApplicationsdefaultJSONResponse{
			Body:       company.TsunaguWorksApiError{Error: "一覧の取得に失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	items := make([]company.TsunaguWorksApplicationForCompany, len(rows))
	for i, r := range rows {
		items[i] = toAPIApplicationForCompany(r.ID, r.ProjectID, r.Status, r.Message, r.TalentDisplayName, r.TalentSkills, r.CreatedAt)
	}
	return company.ProjectsListApplications200JSONResponse{Applications: items}, nil
}

func (h *Handler) ApplicationsOffer(ctx context.Context, req company.ApplicationsOfferRequestObject) (company.ApplicationsOfferResponseObject, error) {
	row, failure := h.changeApplicationStatus(ctx, req.Id, h.application.Offer)
	if failure != nil {
		return company.ApplicationsOfferdefaultJSONResponse(*failure), nil
	}
	return company.ApplicationsOffer200JSONResponse(row), nil
}

func (h *Handler) ApplicationsReject(ctx context.Context, req company.ApplicationsRejectRequestObject) (company.ApplicationsRejectResponseObject, error) {
	row, failure := h.changeApplicationStatus(ctx, req.Id, h.application.Reject)
	if failure != nil {
		return company.ApplicationsRejectdefaultJSONResponse(*failure), nil
	}
	return company.ApplicationsReject200JSONResponse(row), nil
}

// changeApplicationStatus は offer / reject に共通のエラー変換（#42 の changeProjectStatus と同じ型）
func (h *Handler) changeApplicationStatus(
	ctx context.Context,
	applicationID int64,
	action func(context.Context, int64, int64) (dbApplicationRow, error),
) (company.TsunaguWorksApplicationForCompany, *failureResponse) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return company.TsunaguWorksApplicationForCompany{}, &failureResponse{
			Body:       company.TsunaguWorksApiError{Error: "認証が必要です"},
			StatusCode: http.StatusUnauthorized,
		}
	}

	row, err := action(ctx, claims.UserID, applicationID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrApplicationNotFound):
			return company.TsunaguWorksApplicationForCompany{}, &failureResponse{
				Body:       company.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusNotFound,
			}
		case errors.Is(err, usecase.ErrCannotChangeApplication):
			return company.TsunaguWorksApplicationForCompany{}, &failureResponse{
				Body:       company.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusConflict,
			}
		case errors.Is(err, usecase.ErrAuthFailed):
			return company.TsunaguWorksApplicationForCompany{}, &failureResponse{
				Body:       company.TsunaguWorksApiError{Error: "認証が必要です"},
				StatusCode: http.StatusUnauthorized,
			}
		}
		slog.ErrorContext(ctx, "company changeApplicationStatus", "err", err)
		return company.TsunaguWorksApplicationForCompany{}, &failureResponse{
			Body:       company.TsunaguWorksApiError{Error: "選考の操作に失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}
	}

	return toAPIApplicationForCompany(row.ID, row.ProjectID, row.Status, row.Message, row.TalentDisplayName, row.TalentSkills, row.CreatedAt), nil
}

// offer / reject の default レスポンスは同じ形なので、共通処理では素の構造体で持ち、
// 呼び出し側が生成型（ApplicationsOffer/RejectdefaultJSONResponse）へ変換する
type failureResponse = struct {
	Body       company.TsunaguWorksApiError
	StatusCode int
}

// dbApplicationRow は選考遷移クエリの戻り行（usecase の Offer / Reject が返す型）
type dbApplicationRow = db.UpdateApplicationStatusForCompanyRow

func toAPIApplicationForCompany(id, projectID int64, status, message, displayName string, skills []string, createdAt time.Time) company.TsunaguWorksApplicationForCompany {
	return company.TsunaguWorksApplicationForCompany{
		Id:                id,
		ProjectId:         projectID,
		Status:            company.TsunaguWorksApplicationStatus(status),
		Message:           message,
		TalentDisplayName: displayName,
		TalentSkills:      skills,
		CreatedAt:         createdAt,
	}
}
