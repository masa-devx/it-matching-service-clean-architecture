-- 応募（talent → project の交差テーブル）。状態機械②の土台。
-- status の6値は最初から全部 CHECK に入れる（状態機械の全体像がスキーマから読める・遷移実装時にDDL変更不要）。
-- 時刻列の NULL は「未対応」を表す（時刻に自然なゼロ値が無いため、NOT NULL 方針の例外とする）

-- +migrate Up
CREATE TABLE applications (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    talent_id BIGINT NOT NULL REFERENCES talents (id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'applied'
        CHECK (status IN ('applied', 'offered', 'accepted', 'rejected', 'withdrawn', 'declined')),
    message TEXT NOT NULL DEFAULT '',
    company_acted_at TIMESTAMPTZ,
    talent_acted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- 二重応募を DB が物理的に拒否する（アプリ側チェックだけでは同時リクエストで破れる）
    UNIQUE (project_id, talent_id)
);

-- UNIQUE (project_id, talent_id) の索引は左端規則により talent_id 単独の検索に効かないため、
-- 「自分の応募一覧」用に別途張る
CREATE INDEX idx_applications_talent_id ON applications (talent_id);

-- +migrate Down
DROP TABLE applications;
