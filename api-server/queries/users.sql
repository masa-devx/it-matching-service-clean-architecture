-- 認証情報のクエリ。password_hash を返すのは認証フロー（login の照合）に必要なため。
-- API のレスポンスには絶対に載せない（handler の詰め替えで落とす）

-- name: CreateUser :one
INSERT INTO users (email, password_hash, role)
VALUES ($1, $2, $3)
RETURNING id, email, password_hash, role, created_at;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, role, created_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, email, password_hash, role, created_at
FROM users
WHERE id = $1;
