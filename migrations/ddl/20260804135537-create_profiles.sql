-- +migrate Up
-- 企業プロフィール（users と 1:1）
-- user_id の UNIQUE が「1ユーザー1プロフィール」を保証し、外部キーの索引も兼ねる
CREATE TABLE companies (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT      NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    name        TEXT        NOT NULL,
    -- 未入力は NULL ではなく空値で表現する（Go側のポインタ地獄と NULL 比較の罠を避ける）
    description TEXT        NOT NULL DEFAULT '',
    industry    TEXT        NOT NULL DEFAULT '',
    size        TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 人材プロフィール（users と 1:1）
CREATE TABLE talents (
    id                       BIGSERIAL PRIMARY KEY,
    user_id                  BIGINT      NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    bio                      TEXT        NOT NULL DEFAULT '',
    -- スキルは配列型。後の案件検索で「Go を含む人材」を GIN インデックスで引く
    skills                   TEXT[]      NOT NULL DEFAULT '{}',
    years_of_exp             INT         NOT NULL DEFAULT 0,
    available_hours_per_week INT         NOT NULL DEFAULT 0,
    desired_hourly_rate      INT         NOT NULL DEFAULT 0,
    remote_ok                BOOLEAN     NOT NULL DEFAULT false,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 配列の包含検索（skills @> '{Go}'）を高速化する。B-tree では配列の中身を引けない
CREATE INDEX talents_skills_idx ON talents USING GIN (skills);

-- 列の意味を DB 自身に持たせる（DBeaver や \d+ で参照できるデータ辞書）。
-- 名前だけでは意味・単位・区分が分からない列にのみ付ける
COMMENT ON COLUMN companies.user_id IS 'users との1:1関連（UNIQUE制約で1ユーザー1件）';
COMMENT ON COLUMN companies.size IS '従業員規模の区分（例: 1-10 / 11-50 / 51-200 / 201+）';

COMMENT ON COLUMN talents.user_id IS 'users との1:1関連（UNIQUE制約で1ユーザー1件）';
COMMENT ON COLUMN talents.skills IS 'スキルタグの配列（例: {Go,React,AWS}）。GINインデックスで包含検索する';
COMMENT ON COLUMN talents.years_of_exp IS '実務経験年数（0=未設定）';
COMMENT ON COLUMN talents.available_hours_per_week IS '週あたりの稼働可能時間（0=未設定）';
COMMENT ON COLUMN talents.desired_hourly_rate IS '希望時給・円（0=未設定）';
COMMENT ON COLUMN talents.remote_ok IS 'フルリモート勤務の可否';

-- +migrate Down
DROP TABLE talents;
DROP TABLE companies;
