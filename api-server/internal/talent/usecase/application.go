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
	// ErrCannotWithdraw は決着済み等、遷移表が取り下げを許可しない状態（409）
	ErrCannotWithdraw = errors.New("現在の状態では取り下げできません")
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

// withdrawableFrom は「talent が withdrawn へ遷移できる元状態」を遷移表から導出する。
// SQL にハードコードしない＝遷移表（shared/domain）が一次情報のまま WHERE に反映される
func withdrawableFrom() []string {
	var froms []string
	for _, from := range domain.AllApplicationStatuses {
		if domain.CanTransitApplication(domain.ActorTalent, from, domain.ApplicationWithdrawn) {
			froms = append(froms, string(from))
		}
	}
	return froms
}

// Withdraw は応募を取り下げる。所有（talent_id）と遷移可否（from_statuses）は
// 条件付き UPDATE が原子的に検査し、0行なら再取得して 404 と 409 を区別する
func (u *Application) Withdraw(ctx context.Context, userID, applicationID int64) (db.WithdrawApplicationRow, error) {
	talentID, err := u.talentIDFor(ctx, userID)
	if err != nil {
		return db.WithdrawApplicationRow{}, err
	}

	row, err := u.queries.WithdrawApplication(ctx, db.WithdrawApplicationParams{
		ID:           applicationID,
		TalentID:     talentID,
		FromStatuses: withdrawableFrom(),
	})
	if err == nil {
		return row, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.WithdrawApplicationRow{}, fmt.Errorf("取り下げに失敗: %w", err)
	}

	// 0行 = 不存在/他人（404）か、遷移不可の状態（409）かを取得して区別する
	app, getErr := u.queries.GetApplicationForTalent(ctx, db.GetApplicationForTalentParams{
		ID:       applicationID,
		TalentID: talentID,
	})
	if getErr != nil {
		if errors.Is(getErr, pgx.ErrNoRows) {
			return db.WithdrawApplicationRow{}, ErrApplicationNotFound
		}
		return db.WithdrawApplicationRow{}, fmt.Errorf("応募の取得に失敗: %w", getErr)
	}
	return db.WithdrawApplicationRow{}, fmt.Errorf("%w（現在: %s）", ErrCannotWithdraw, app.Status)
}
