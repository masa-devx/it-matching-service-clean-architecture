-- 人材プロフィールのクエリ

-- name: CreateTalent :one
INSERT INTO talents (user_id, display_name, skills, bio)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, display_name, skills, bio, created_at;

-- name: GetTalentByUserID :one
SELECT id, user_id, display_name, skills, bio, created_at
FROM talents
WHERE user_id = $1;
