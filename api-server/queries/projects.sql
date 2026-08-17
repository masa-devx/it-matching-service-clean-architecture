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
