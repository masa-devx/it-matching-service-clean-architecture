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
)

func (h *Handler) ApplicationsCreate(ctx context.Context, req talent.ApplicationsCreateRequestObject) (talent.ApplicationsCreateResponseObject, error) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return talent.ApplicationsCreatedefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "認証が必要です"},
			StatusCode: http.StatusUnauthorized,
		}, nil
	}
	if req.Body == nil {
		return talent.ApplicationsCreatedefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "リクエストボディが必要です"},
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	row, err := h.application.Apply(ctx, claims.UserID, req.Body.ProjectId, deref(req.Body.Message))
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrProjectNotFound):
			return talent.ApplicationsCreatedefaultJSONResponse{
				Body:       talent.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusNotFound,
			}, nil
		case errors.Is(err, usecase.ErrAlreadyApplied):
			return talent.ApplicationsCreatedefaultJSONResponse{
				Body:       talent.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusConflict,
			}, nil
		case errors.Is(err, usecase.ErrAuthFailed):
			return talent.ApplicationsCreatedefaultJSONResponse{
				Body:       talent.TsunaguWorksApiError{Error: "認証が必要です"},
				StatusCode: http.StatusUnauthorized,
			}, nil
		}
		log.Printf("talent ApplicationsCreate: %v", err)
		return talent.ApplicationsCreatedefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "応募に失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return talent.ApplicationsCreate201JSONResponse(toAPIApplication(
		row.ID, row.ProjectID, row.ProjectTitle, row.Status, row.Message, row.CreatedAt,
	)), nil
}

func (h *Handler) ApplicationsList(ctx context.Context, req talent.ApplicationsListRequestObject) (talent.ApplicationsListResponseObject, error) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return talent.ApplicationsListdefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "認証が必要です"},
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	rows, err := h.application.ListMine(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, usecase.ErrAuthFailed) {
			return talent.ApplicationsListdefaultJSONResponse{
				Body:       talent.TsunaguWorksApiError{Error: "認証が必要です"},
				StatusCode: http.StatusUnauthorized,
			}, nil
		}
		log.Printf("talent ApplicationsList: %v", err)
		return talent.ApplicationsListdefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "一覧の取得に失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	items := make([]talent.TsunaguWorksApplication, len(rows))
	for i, r := range rows {
		items[i] = toAPIApplication(r.ID, r.ProjectID, r.ProjectTitle, r.Status, r.Message, r.CreatedAt)
	}
	return talent.ApplicationsList200JSONResponse{Applications: items}, nil
}

func (h *Handler) ApplicationsWithdraw(ctx context.Context, req talent.ApplicationsWithdrawRequestObject) (talent.ApplicationsWithdrawResponseObject, error) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return talent.ApplicationsWithdrawdefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "認証が必要です"},
			StatusCode: http.StatusUnauthorized,
		}, nil
	}

	row, err := h.application.Withdraw(ctx, claims.UserID, req.Id)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrApplicationNotFound):
			return talent.ApplicationsWithdrawdefaultJSONResponse{
				Body:       talent.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusNotFound,
			}, nil
		case errors.Is(err, usecase.ErrCannotChangeApplication):
			// 遷移不可の詳細（現在の状態）はラップ済みメッセージに含まれている
			return talent.ApplicationsWithdrawdefaultJSONResponse{
				Body:       talent.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusConflict,
			}, nil
		case errors.Is(err, usecase.ErrAuthFailed):
			return talent.ApplicationsWithdrawdefaultJSONResponse{
				Body:       talent.TsunaguWorksApiError{Error: "認証が必要です"},
				StatusCode: http.StatusUnauthorized,
			}, nil
		}
		log.Printf("talent ApplicationsWithdraw: %v", err)
		return talent.ApplicationsWithdrawdefaultJSONResponse{
			Body:       talent.TsunaguWorksApiError{Error: "取り下げに失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	return talent.ApplicationsWithdraw200JSONResponse(toAPIApplication(
		row.ID, row.ProjectID, row.ProjectTitle, row.Status, row.Message, row.CreatedAt,
	)), nil
}

// 承諾・辞退のスタブ（#58 Step 2 で withdraw ごと共通ヘルパーに置き換える）

func (h *Handler) ApplicationsAccept(ctx context.Context, req talent.ApplicationsAcceptRequestObject) (talent.ApplicationsAcceptResponseObject, error) {
	return talent.ApplicationsAcceptdefaultJSONResponse{
		Body:       talent.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}

func (h *Handler) ApplicationsDecline(ctx context.Context, req talent.ApplicationsDeclineRequestObject) (talent.ApplicationsDeclineResponseObject, error) {
	return talent.ApplicationsDeclinedefaultJSONResponse{
		Body:       talent.TsunaguWorksApiError{Error: "未実装です"},
		StatusCode: http.StatusNotImplemented,
	}, nil
}

// toAPIApplication は DB の行（JOIN 済み Row 各種）を API の型へ詰め替える。
// sqlc がクエリごとに別の Row 型を生成するため、共通のフィールドを引数で受ける
func toAPIApplication(id, projectID int64, title, status, message string, createdAt time.Time) talent.TsunaguWorksApplication {
	return talent.TsunaguWorksApplication{
		Id:           id,
		ProjectId:    projectID,
		ProjectTitle: title,
		Status:       talent.TsunaguWorksApplicationStatus(status),
		Message:      message,
		CreatedAt:    createdAt,
	}
}
