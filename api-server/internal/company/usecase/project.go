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

// Create は案件を draft として作成する。
// 所有者（company_id）はクライアントから受け取らず、検証済みトークンの userID から
// プロフィールを引いて決める——「他社を所有者に指定する」形が存在しない（IDOR対策）
func (u *Project) Create(ctx context.Context, userID int64, params db.CreateProjectParams) (db.Project, error) {
	comp, err := u.queries.GetCompanyByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// トークンは有効だがプロフィールの実体が無い（削除済み等）
			return db.Project{}, ErrAuthFailed
		}
		return db.Project{}, fmt.Errorf("企業プロフィール取得に失敗: %w", err)
	}
	params.CompanyID = comp.ID

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

// ChangeStatus は掲載状態を遷移させる。可否の判断は shared/domain の遷移表に委ねる
// （usecase は「取得→表に聞く→条件付き更新」の段取りだけを持つ）
func (u *Project) ChangeStatus(ctx context.Context, userID, projectID int64, to domain.ProjectStatus) (db.Project, error) {
	comp, err := u.queries.GetCompanyByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Project{}, ErrAuthFailed
		}
		return db.Project{}, fmt.Errorf("企業プロフィール取得に失敗: %w", err)
	}

	proj, err := u.queries.GetProjectForCompany(ctx, db.GetProjectForCompanyParams{
		ID:        projectID,
		CompanyID: comp.ID,
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
		CompanyID:  comp.ID,
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
