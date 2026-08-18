package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/domain"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/infra"
)

var (
	// ErrApplicationNotFound は不存在と他人の応募を区別しない（存在の有無を漏らさない・404）
	ErrApplicationNotFound = errors.New("応募が見つかりません")
	// ErrAlreadyApplied は二重応募（UNIQUE 制約違反・409）
	ErrAlreadyApplied = errors.New("この案件にはすでに応募しています")
	// ErrCannotChangeApplication は決着済み等、遷移表が許可しない状態への操作（409）
	ErrCannotChangeApplication = errors.New("現在の状態ではこの操作はできません")
)

// Application は応募（talent 視点: 出す・取り下げる）の業務ロジック
type Application struct {
	queries *db.Queries
}

func NewApplication(queries *db.Queries) *Application {
	return &Application{queries: queries}
}

// talentIDFor は検証済みトークンの userID から人材プロフィール ID を解決する。
// 応募者はすべての操作でこの経路からしか決まらない（company 側 companyIDFor と対称・IDOR対策の一本化）
func (u *Application) talentIDFor(ctx context.Context, userID int64) (int64, error) {
	tal, err := u.queries.GetTalentByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrAuthFailed
		}
		return 0, fmt.Errorf("人材プロフィール取得に失敗: %w", err)
	}
	return tal.ID, nil
}

// Apply は公開中の案件に応募する。
// 「公開中のみ」の検査は INSERT...SELECT が原子的に行う（0行 = 未公開 or 不存在 = 404）。
// 二重応募は UNIQUE(project_id, talent_id) 違反として DB が拒否する（409）
func (u *Application) Apply(ctx context.Context, userID, projectID int64, message string) (db.GetApplicationForTalentRow, error) {
	talentID, err := u.talentIDFor(ctx, userID)
	if err != nil {
		return db.GetApplicationForTalentRow{}, err
	}

	app, err := u.queries.CreateApplication(ctx, db.CreateApplicationParams{
		TalentID:  talentID,
		Message:   message,
		ProjectID: projectID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetApplicationForTalentRow{}, ErrProjectNotFound
		}
		if infra.IsUniqueViolation(err) {
			return db.GetApplicationForTalentRow{}, ErrAlreadyApplied
		}
		return db.GetApplicationForTalentRow{}, fmt.Errorf("応募の作成に失敗: %w", err)
	}

	// レスポンス形（project_title 込み）は取得クエリで組み立てる
	row, err := u.queries.GetApplicationForTalent(ctx, db.GetApplicationForTalentParams{
		ID:       app.ID,
		TalentID: talentID,
	})
	if err != nil {
		return db.GetApplicationForTalentRow{}, fmt.Errorf("作成した応募の取得に失敗: %w", err)
	}
	return row, nil
}

// ListMine は自分の応募一覧を返す（新しい順・案件タイトル込み）
func (u *Application) ListMine(ctx context.Context, userID int64) ([]db.ListApplicationsForTalentRow, error) {
	talentID, err := u.talentIDFor(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows, err := u.queries.ListApplicationsForTalent(ctx, talentID)
	if err != nil {
		return nil, fmt.Errorf("応募一覧の取得に失敗: %w", err)
	}
	return rows, nil
}

// Withdraw は応募を取り下げる（applied / offered → withdrawn）
func (u *Application) Withdraw(ctx context.Context, userID, applicationID int64) (db.UpdateApplicationStatusForTalentRow, error) {
	return u.changeStatus(ctx, userID, applicationID, domain.ApplicationWithdrawn)
}

// Accept はオファーを承諾する（offered → accepted・ダブルオプトインの成立）
func (u *Application) Accept(ctx context.Context, userID, applicationID int64) (db.UpdateApplicationStatusForTalentRow, error) {
	return u.changeStatus(ctx, userID, applicationID, domain.ApplicationAccepted)
}

// Decline はオファーを辞退する（offered → declined）
func (u *Application) Decline(ctx context.Context, userID, applicationID int64) (db.UpdateApplicationStatusForTalentRow, error) {
	return u.changeStatus(ctx, userID, applicationID, domain.ApplicationDeclined)
}

// changeableFrom は「talent が to へ遷移させられる元状態」を遷移表から導出する。
// SQL にハードコードしない＝遷移表（shared/domain）が一次情報のまま WHERE に反映される
func changeableFrom(to domain.ApplicationStatus) []string {
	var froms []string
	for _, from := range domain.AllApplicationStatuses {
		if domain.CanTransitApplication(domain.ActorTalent, from, to) {
			froms = append(froms, string(from))
		}
	}
	return froms
}

// changeStatus は talent の遷移を実行する（withdraw / accept / decline の共通実体）。
// 所有（talent_id）と遷移可否（from_statuses）は条件付き UPDATE が原子的に検査し、
// 0行なら再取得して 404 と 409 を区別する
func (u *Application) changeStatus(ctx context.Context, userID, applicationID int64, to domain.ApplicationStatus) (db.UpdateApplicationStatusForTalentRow, error) {
	talentID, err := u.talentIDFor(ctx, userID)
	if err != nil {
		return db.UpdateApplicationStatusForTalentRow{}, err
	}

	row, err := u.queries.UpdateApplicationStatusForTalent(ctx, db.UpdateApplicationStatusForTalentParams{
		ID:           applicationID,
		TalentID:     talentID,
		ToStatus:     string(to),
		FromStatuses: changeableFrom(to),
	})
	if err == nil {
		return row, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.UpdateApplicationStatusForTalentRow{}, fmt.Errorf("応募状態の更新に失敗: %w", err)
	}

	// 0行 = 不存在/他人（404）か、遷移不可の状態（409）かを取得して区別する
	app, getErr := u.queries.GetApplicationForTalent(ctx, db.GetApplicationForTalentParams{
		ID:       applicationID,
		TalentID: talentID,
	})
	if getErr != nil {
		if errors.Is(getErr, pgx.ErrNoRows) {
			return db.UpdateApplicationStatusForTalentRow{}, ErrApplicationNotFound
		}
		return db.UpdateApplicationStatusForTalentRow{}, fmt.Errorf("応募の取得に失敗: %w", getErr)
	}
	return db.UpdateApplicationStatusForTalentRow{}, fmt.Errorf("%w（現在: %s）", ErrCannotChangeApplication, app.Status)
}
