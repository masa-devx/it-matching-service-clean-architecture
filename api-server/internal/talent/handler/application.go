package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	talent "github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/api/talent"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
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
		slog.ErrorContext(ctx, "talent ApplicationsCreate", "err", err)
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
		slog.ErrorContext(ctx, "talent ApplicationsList", "err", err)
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
	app, failure := h.changeApplicationStatus(ctx, req.Id, h.application.Withdraw)
	if failure != nil {
		return talent.ApplicationsWithdrawdefaultJSONResponse(*failure), nil
	}
	return talent.ApplicationsWithdraw200JSONResponse(app), nil
}

func (h *Handler) ApplicationsAccept(ctx context.Context, req talent.ApplicationsAcceptRequestObject) (talent.ApplicationsAcceptResponseObject, error) {
	app, failure := h.changeApplicationStatus(ctx, req.Id, h.application.Accept)
	if failure != nil {
		return talent.ApplicationsAcceptdefaultJSONResponse(*failure), nil
	}
	return talent.ApplicationsAccept200JSONResponse(app), nil
}

func (h *Handler) ApplicationsDecline(ctx context.Context, req talent.ApplicationsDeclineRequestObject) (talent.ApplicationsDeclineResponseObject, error) {
	app, failure := h.changeApplicationStatus(ctx, req.Id, h.application.Decline)
	if failure != nil {
		return talent.ApplicationsDeclinedefaultJSONResponse(*failure), nil
	}
	return talent.ApplicationsDecline200JSONResponse(app), nil
}

// withdraw / accept / decline の default レスポンスは同じ形なので、共通処理では素の構造体で持ち、
// 呼び出し側が生成型へ変換する（company 側 changeApplicationStatus と同じ型・#57）
type failureResponse = struct {
	Body       talent.TsunaguWorksApiError
	StatusCode int
}

// changeApplicationStatus は3操作に共通のエラー変換。usecase のメソッド値を action として受ける
func (h *Handler) changeApplicationStatus(
	ctx context.Context,
	applicationID int64,
	action func(context.Context, int64, int64) (db.UpdateApplicationStatusForTalentRow, error),
) (talent.TsunaguWorksApplication, *failureResponse) {
	claims, ok := auth.ClaimsFrom(ctx)
	if !ok {
		return talent.TsunaguWorksApplication{}, &failureResponse{
			Body:       talent.TsunaguWorksApiError{Error: "認証が必要です"},
			StatusCode: http.StatusUnauthorized,
		}
	}

	row, err := action(ctx, claims.UserID, applicationID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrApplicationNotFound):
			return talent.TsunaguWorksApplication{}, &failureResponse{
				Body:       talent.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusNotFound,
			}
		case errors.Is(err, usecase.ErrCannotChangeApplication):
			// 遷移不可の詳細（現在の状態）はラップ済みメッセージに含まれている
			return talent.TsunaguWorksApplication{}, &failureResponse{
				Body:       talent.TsunaguWorksApiError{Error: err.Error()},
				StatusCode: http.StatusConflict,
			}
		case errors.Is(err, usecase.ErrAuthFailed):
			return talent.TsunaguWorksApplication{}, &failureResponse{
				Body:       talent.TsunaguWorksApiError{Error: "認証が必要です"},
				StatusCode: http.StatusUnauthorized,
			}
		}
		slog.ErrorContext(ctx, "talent changeApplicationStatus", "err", err)
		return talent.TsunaguWorksApplication{}, &failureResponse{
			Body:       talent.TsunaguWorksApiError{Error: "応募の操作に失敗しました"},
			StatusCode: http.StatusInternalServerError,
		}
	}

	return toAPIApplication(row.ID, row.ProjectID, row.ProjectTitle, row.Status, row.Message, row.CreatedAt), nil
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
