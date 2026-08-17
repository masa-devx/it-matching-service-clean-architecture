-- 認証情報のテーブル。プロフィール（companies / talents）とは分離する:
-- 「ログインの仕組み」と「画面に見える情報」は変更理由が違うため

-- +migrate Up
CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('company', 'talent')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +migrate Down
DROP TABLE users;
