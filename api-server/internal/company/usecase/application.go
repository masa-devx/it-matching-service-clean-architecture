package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/domain"
)

var (
	// ErrApplicationNotFound は不存在と他社案件の応募を区別しない（存在の有無を漏らさない・404）
	ErrApplicationNotFound = errors.New("応募が見つかりません")
	// ErrCannotChangeApplication は遷移表が許可しない状態への操作（409）
	ErrCannotChangeApplication = errors.New("現在の状態ではこの操作はできません")
)

// Application は選考（company 視点: 受け取る・選考する）の業務ロジック
type Application struct {
	queries *db.Queries
}

func NewApplication(queries *db.Queries) *Application {
	return &Application{queries: queries}
}

// ListForProject は自社案件に届いた応募の一覧を返す。
// 所有確認（GetProjectForCompany）を先に行うのは、「他社の案件（404）」と
// 「自社の案件で応募ゼロ（空リスト）」を区別するため。
// WHERE に埋め込む1クエリ方式だと両者が同じ0行になり、存在の有無が漏れる
func (u *Application) ListForProject(ctx context.Context, userID, projectID int64) ([]db.ListApplicationsForProjectRow, error) {
	companyID, err := resolveCompanyID(ctx, u.queries, userID)
	if err != nil {
		return nil, err
	}

	if _, err := u.queries.GetProjectForCompany(ctx, db.GetProjectForCompanyParams{
		ID:        projectID,
		CompanyID: companyID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("案件の取得に失敗: %w", err)
	}

	rows, err := u.queries.ListApplicationsForProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("応募一覧の取得に失敗: %w", err)
	}
	return rows, nil
}

// Offer はオファーする（applied → offered）。message は応募者への一言（空なら「無し」= NULL で保存）
func (u *Application) Offer(ctx context.Context, userID, applicationID int64, message string) (db.UpdateApplicationStatusForCompanyRow, error) {
	var offerMessage *string
	if message != "" {
		offerMessage = &message
	}
	return u.changeStatus(ctx, userID, applicationID, domain.ApplicationOffered, offerMessage)
}

// Reject は不採用にする（applied → rejected）。offer_message は設定しない（NULL のまま）
func (u *Application) Reject(ctx context.Context, userID, applicationID int64) (db.UpdateApplicationStatusForCompanyRow, error) {
	return u.changeStatus(ctx, userID, applicationID, domain.ApplicationRejected, nil)
}

// changeableFrom は「company が to へ遷移させられる元状態」を遷移表から導出する
// （talent 側 withdrawableFrom と同じ型。表が一次情報のまま WHERE に反映される）
func changeableFrom(to domain.ApplicationStatus) []string {
	var froms []string
	for _, from := range domain.AllApplicationStatuses {
		if domain.CanTransitApplication(domain.ActorCompany, from, to) {
			froms = append(froms, string(from))
		}
	}
	return froms
}

// changeStatus は選考の遷移を実行する。所有（JOIN 越しの company_id）と遷移可否は
// 条件付き UPDATE が原子的に検査し、0行なら再取得して 404 と 409 を区別する
func (u *Application) changeStatus(ctx context.Context, userID, applicationID int64, to domain.ApplicationStatus, offerMessage *string) (db.UpdateApplicationStatusForCompanyRow, error) {
	companyID, err := resolveCompanyID(ctx, u.queries, userID)
	if err != nil {
		return db.UpdateApplicationStatusForCompanyRow{}, err
	}

	row, err := u.queries.UpdateApplicationStatusForCompany(ctx, db.UpdateApplicationStatusForCompanyParams{
		ID:           applicationID,
		CompanyID:    companyID,
		ToStatus:     string(to),
		FromStatuses: changeableFrom(to),
		OfferMessage: offerMessage,
	})
	if err == nil {
		return row, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.UpdateApplicationStatusForCompanyRow{}, fmt.Errorf("選考状態の更新に失敗: %w", err)
	}

	app, getErr := u.queries.GetApplicationForCompany(ctx, db.GetApplicationForCompanyParams{
		ID:        applicationID,
		CompanyID: companyID,
	})
	if getErr != nil {
		if errors.Is(getErr, pgx.ErrNoRows) {
			return db.UpdateApplicationStatusForCompanyRow{}, ErrApplicationNotFound
		}
		return db.UpdateApplicationStatusForCompanyRow{}, fmt.Errorf("応募の取得に失敗: %w", getErr)
	}
	return db.UpdateApplicationStatusForCompanyRow{}, fmt.Errorf("%w（現在: %s）", ErrCannotChangeApplication, app.Status)
}
