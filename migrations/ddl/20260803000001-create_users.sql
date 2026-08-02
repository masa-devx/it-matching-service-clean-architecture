-- +migrate Up
-- ユーザー（企業・人材の共通アカウント）
-- role で企業(company)/人材(talent)を区別する。詳細プロフィールは
-- 後続の companies / talents テーブル（1:1）に分離する予定
CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT        NOT NULL UNIQUE,
    -- bcrypt ハッシュを保存する。平文パスワードは絶対に保存しない
    password_hash TEXT        NOT NULL,
    role          TEXT        NOT NULL CHECK (role IN ('company', 'talent')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +migrate Down
DROP TABLE users;
