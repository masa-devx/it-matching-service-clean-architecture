-- 企業プロフィールのクエリ

-- name: CreateCompany :one
INSERT INTO companies (user_id, name, location, description)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, name, location, description, created_at;

-- name: GetCompanyByUserID :one
SELECT id, user_id, name, location, description, created_at
FROM companies
WHERE user_id = $1;
