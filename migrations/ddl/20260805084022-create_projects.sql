-- +migrate Up
-- 案件（企業が掲載する仕事）。companies と 1:N（1社が複数案件を持つ）
CREATE TABLE projects (
    id              BIGSERIAL PRIMARY KEY,
    -- 1:N なので UNIQUE は付けない。外部キーの索引は別途明示的に張る（下部の CREATE INDEX）
    company_id      BIGINT      NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    title           TEXT        NOT NULL,
    description     TEXT        NOT NULL DEFAULT '',
    required_skills TEXT[]      NOT NULL DEFAULT '{}',
    hourly_rate_min INT         NOT NULL DEFAULT 0,
    hourly_rate_max INT         NOT NULL DEFAULT 0,
    hours_per_week  INT         NOT NULL DEFAULT 0,
    remote_ok       BOOLEAN     NOT NULL DEFAULT false,
    -- 掲載の状態。公開前の下書き・募集終了を区別する（状態遷移の第一歩）
    status          TEXT        NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft', 'published', 'closed')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- 単価は「下限 <= 上限」でなければ意味を成さない。DBでも保証する
    CONSTRAINT projects_rate_range_check CHECK (hourly_rate_min <= hourly_rate_max)
);

-- 「この会社の案件一覧」を引くための索引（外部キーには索引が必要）
CREATE INDEX projects_company_id_idx ON projects (company_id);

-- 一覧は「公開中のものを新しい順」で引く。複合索引で並べ替えごと引けるようにする
CREATE INDEX projects_status_created_at_idx ON projects (status, created_at DESC);

-- スキルの包含検索（required_skills @> '{Go}'）用
CREATE INDEX projects_required_skills_idx ON projects USING GIN (required_skills);

COMMENT ON COLUMN projects.company_id IS '掲載元の企業（companies.id）。1社が複数案件を持つ 1:N';
COMMENT ON COLUMN projects.required_skills IS '必須スキルの配列（例: {Go,React}）。GINインデックスで包含検索する';
COMMENT ON COLUMN projects.hourly_rate_min IS '時給の下限・円（0=未設定）';
COMMENT ON COLUMN projects.hourly_rate_max IS '時給の上限・円（0=未設定）';
COMMENT ON COLUMN projects.hours_per_week IS '想定稼働時間・週あたり（0=未設定）';
COMMENT ON COLUMN projects.remote_ok IS 'フルリモート可否';
COMMENT ON COLUMN projects.status IS '掲載状態: draft=下書き / published=公開中 / closed=募集終了';

-- +migrate Down
DROP TABLE projects;
