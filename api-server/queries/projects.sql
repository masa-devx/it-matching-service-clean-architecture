-- 案件のクエリ。status は列挙しない（DB の DEFAULT 'draft' に任せる＝作成時に公開状態を指定できる形を作らない）

-- name: CreateProject :one
INSERT INTO projects (
    title,
    description,
    hourly_rate_min,
    hourly_rate_max,
    hours_per_week,
    remote_ok,
    required_skills
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
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
