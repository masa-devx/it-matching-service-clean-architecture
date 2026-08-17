-- 案件のクエリ。status は列挙しない（DB の DEFAULT 'draft' に任せる＝作成時に公開状態を指定できる形を作らない）

-- name: CreateProject :one
INSERT INTO projects (
    company_id,
    title,
    description,
    hourly_rate_min,
    hourly_rate_max,
    hours_per_week,
    remote_ok,
    required_skills
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING
    id,
    title,
    description,
    hourly_rate_min,
    hourly_rate_max,
    hours_per_week,
    remote_ok,
    required_skills,
    status,
    created_at,
    company_id;

-- name: GetProjectForCompany :one
-- 所有チェック込みの取得。WHERE に company_id を含める＝他社の案件は「存在しない」のと同じ扱い
SELECT
    id,
    title,
    description,
    hourly_rate_min,
    hourly_rate_max,
    hours_per_week,
    remote_ok,
    required_skills,
    status,
    created_at,
    company_id
FROM projects
WHERE id = $1 AND company_id = $2;

-- name: ListProjectsForCompany :many
-- 自社の案件のみ（下書き含む・新しい順）。WHERE company_id ＝ 他社の案件は一覧に存在しない
SELECT
    id,
    title,
    description,
    hourly_rate_min,
    hourly_rate_max,
    hours_per_week,
    remote_ok,
    required_skills,
    status,
    created_at,
    company_id
FROM projects
WHERE company_id = $1
ORDER BY id DESC;

-- name: UpdateProject :one
-- 編集可能な列だけを SET する。status / company_id は SQL 文に存在しない
-- ＝どんな入力が来ても「編集で公開状態が変わる」「所有者が変わる」形が無い（1クエリ1意図）
UPDATE projects
SET
    title = $3,
    description = $4,
    hourly_rate_min = $5,
    hourly_rate_max = $6,
    hours_per_week = $7,
    remote_ok = $8,
    required_skills = $9
WHERE id = $1 AND company_id = $2
RETURNING
    id,
    title,
    description,
    hourly_rate_min,
    hourly_rate_max,
    hours_per_week,
    remote_ok,
    required_skills,
    status,
    created_at,
    company_id;

-- name: ListPublishedProjects :many
-- talent 向けの公開案件一覧。seek 法: OFFSET を使わず「id という不動の座標」から続きを読む
-- （挿入があってもページがズレず、深いページでも速い）。limit+1 件の取得は
-- 呼び出し側の「次ページ有無」判定用（n+1 テクニック）。
-- 検索条件は narg（NULL＝条件を無効化）で SQL 1本のまま動的にする
SELECT
    id,
    title,
    description,
    hourly_rate_min,
    hourly_rate_max,
    hours_per_week,
    remote_ok,
    required_skills,
    status,
    created_at,
    company_id
FROM projects
WHERE status = 'published'
    AND (sqlc.narg('cursor')::bigint IS NULL OR id <= sqlc.narg('cursor'))
    AND (sqlc.narg('skills')::text[] IS NULL OR required_skills @> sqlc.narg('skills'))
    AND (sqlc.narg('remote_ok')::boolean IS NULL OR remote_ok = sqlc.narg('remote_ok'))
    AND (sqlc.narg('min_hourly_rate')::int IS NULL OR hourly_rate_min >= sqlc.narg('min_hourly_rate'))
ORDER BY id DESC
LIMIT sqlc.arg('limit_plus_one');

-- name: GetPublishedProject :one
-- 公開中のみ取得。draft / closed は「存在しない」扱い（未公開のものは取得しない原則）
SELECT
    id,
    title,
    description,
    hourly_rate_min,
    hourly_rate_max,
    hours_per_week,
    remote_ok,
    required_skills,
    status,
    created_at,
    company_id
FROM projects
WHERE id = $1 AND status = 'published';

-- name: UpdateProjectStatus :one
-- 状態遷移の更新。WHERE に「今も遷移元のまま」を含めることで、
-- 判定と更新の間に他リクエストが割り込む競合を DB が原子的に検査する（0行更新＝競合）。
-- status が2回登場するため、取り違え防止に名前付きパラメータ（sqlc の @記法）を使う
UPDATE projects
SET status = @to_status
WHERE id = @id AND company_id = @company_id AND status = @from_status
RETURNING
    id,
    title,
    description,
    hourly_rate_min,
    hourly_rate_max,
    hours_per_week,
    remote_ok,
    required_skills,
    status,
    created_at,
    company_id;
