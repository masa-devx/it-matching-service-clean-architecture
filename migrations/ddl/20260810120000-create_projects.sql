-- 案件テーブル（Phase 0: 最小構成。company_id 等の所有関係は認証導入の Phase 2 で追加する）
-- CHECK 制約は packages/spec の制約（@maxLength 等）と対応させる。
-- アプリ側の検証をすり抜けた不正データを DB が最終防衛線として拒否する

-- +migrate Up
CREATE TABLE projects (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title TEXT NOT NULL CHECK (char_length(title) BETWEEN 1 AND 100),
    description TEXT NOT NULL CHECK (char_length(description) BETWEEN 1 AND 5000),
    hourly_rate_min INTEGER CHECK (hourly_rate_min >= 0),
    hourly_rate_max INTEGER CHECK (hourly_rate_max >= 0),
    hours_per_week INTEGER NOT NULL CHECK (hours_per_week BETWEEN 1 AND 60),
    remote_ok BOOLEAN NOT NULL,
    required_skills TEXT[] NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'closed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +migrate Down
DROP TABLE projects;
