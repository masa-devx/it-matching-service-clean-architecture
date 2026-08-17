// Package usecase は company API の業務ロジックを置く。
// 1つの公開メソッドが1つの業務操作を完結させる。複数テーブルへの書き込みが必要になったら、
// この層がトランザクション境界を持つ（Queries.WithTx で束ねる）。
package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/generated/db"
	"github.com/masahiro96848/it-matching-service-clean-architecture/api-server/internal/shared/domain"
)

var (
	// ErrProjectNotFound は不存在と他社所有を区別しない（存在の有無を漏らさない・404）
	ErrProjectNotFound = errors.New("案件が見つかりません")
	// ErrStatusConflict は判定後・更新前に他リクエストが割り込んだ競合（409）
	ErrStatusConflict = errors.New("案件の状態が変更されています。再読み込みしてください")
)

// TransitionError は遷移表で許可されない状態遷移（409）。
// 「今の状態」と「可能な遷移先」を持ち、handler がそのままメッセージにできる
type TransitionError struct {
	From domain.ProjectStatus
	To   domain.ProjectStatus
}

func (e *TransitionError) Error() string {
	allowed := domain.ProjectTransitionsFrom(e.From)
	names := make([]string, len(allowed))
	for i, s := range allowed {
		names[i] = string(s)
	}
	return fmt.Sprintf("%s から %s へは遷移できません（可能な遷移先: %s）", e.From, e.To, strings.Join(names, ", "))
}

// Project は案件の業務ロジック。依存は生成された Queries のみ（HTTP の型を知らない）
type Project struct {
	queries *db.Queries
}

func NewProject(queries *db.Queries) *Project {
	return &Project{queries: queries}
}

// companyIDFor は検証済みトークンの userID から企業プロフィール ID を解決する。
// 所有者はすべての操作でこの経路からしか決まらない（IDOR対策の一本化）
func (u *Project) companyIDFor(ctx context.Context, userID int64) (int64, error) {
	comp, err := u.queries.GetCompanyByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// トークンは有効だがプロフィールの実体が無い（削除済み等）
			return 0, ErrAuthFailed
		}
		return 0, fmt.Errorf("企業プロフィール取得に失敗: %w", err)
	}
	return comp.ID, nil
}

// Create は案件を draft として作成する。
// 所有者（company_id）はクライアントから受け取らず、検証済みトークンの userID から
// プロフィールを引いて決める——「他社を所有者に指定する」形が存在しない（IDOR対策）
func (u *Project) Create(ctx context.Context, userID int64, params db.CreateProjectParams) (db.Project, error) {
	companyID, err := u.companyIDFor(ctx, userID)
	if err != nil {
		return db.Project{}, err
	}
	params.CompanyID = companyID

	// nil スライスは SQL の NULL になり NOT NULL 制約に弾かれる（#30 と同型の正規化）
	if params.RequiredSkills == nil {
		params.RequiredSkills = []string{}
	}

	project, err := u.queries.CreateProject(ctx, params)
	if err != nil {
		return db.Project{}, fmt.Errorf("案件の作成に失敗: %w", err)
	}
	return project, nil
}

// ListMine は自社の案件一覧を返す（下書き含む・新しい順）
func (u *Project) ListMine(ctx context.Context, userID int64) ([]db.Project, error) {
	companyID, err := u.companyIDFor(ctx, userID)
	if err != nil {
		return nil, err
	}
	projects, err := u.queries.ListProjectsForCompany(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("案件一覧の取得に失敗: %w", err)
	}
	return projects, nil
}

// GetMine は自社の案件詳細を返す（他社・不存在は区別せず ErrProjectNotFound）
func (u *Project) GetMine(ctx context.Context, userID, projectID int64) (db.Project, error) {
	companyID, err := u.companyIDFor(ctx, userID)
	if err != nil {
		return db.Project{}, err
	}
	proj, err := u.queries.GetProjectForCompany(ctx, db.GetProjectForCompanyParams{
		ID:        projectID,
		CompanyID: companyID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Project{}, ErrProjectNotFound
		}
		return db.Project{}, fmt.Errorf("案件取得に失敗: %w", err)
	}
	return proj, nil
}

// Update は自社の案件を編集する。status / company_id は SQL の SET 句に存在しないため、
// この操作で公開状態や所有者が変わることは無い
func (u *Project) Update(ctx context.Context, userID, projectID int64, params db.UpdateProjectParams) (db.Project, error) {
	companyID, err := u.companyIDFor(ctx, userID)
	if err != nil {
		return db.Project{}, err
	}
	params.ID = projectID
	params.CompanyID = companyID

	if params.RequiredSkills == nil {
		params.RequiredSkills = []string{}
	}

	updated, err := u.queries.UpdateProject(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Project{}, ErrProjectNotFound
		}
		return db.Project{}, fmt.Errorf("案件の更新に失敗: %w", err)
	}
	return updated, nil
}

// ChangeStatus は掲載状態を遷移させる。可否の判断は shared/domain の遷移表に委ねる
// （usecase は「取得→表に聞く→条件付き更新」の段取りだけを持つ）
func (u *Project) ChangeStatus(ctx context.Context, userID, projectID int64, to domain.ProjectStatus) (db.Project, error) {
	companyID, err := u.companyIDFor(ctx, userID)
	if err != nil {
		return db.Project{}, err
	}

	proj, err := u.queries.GetProjectForCompany(ctx, db.GetProjectForCompanyParams{
		ID:        projectID,
		CompanyID: companyID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Project{}, ErrProjectNotFound
		}
		return db.Project{}, fmt.Errorf("案件取得に失敗: %w", err)
	}

	from := domain.ProjectStatus(proj.Status)
	if !domain.CanTransitProject(from, to) {
		return db.Project{}, &TransitionError{From: from, To: to}
	}

	updated, err := u.queries.UpdateProjectStatus(ctx, db.UpdateProjectStatusParams{
		ID:         projectID,
		CompanyID:  companyID,
		FromStatus: string(from),
		ToStatus:   string(to),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// 判定後に別リクエストが状態を変えた（WHERE の from 条件に外れて 0 行更新）
			return db.Project{}, ErrStatusConflict
		}
		return db.Project{}, fmt.Errorf("状態の更新に失敗: %w", err)
	}
	return updated, nil
}
