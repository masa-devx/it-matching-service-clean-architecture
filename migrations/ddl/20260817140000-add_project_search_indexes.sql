-- talent 向け公開案件一覧のための索引。
-- ① (status, id DESC): seek ページネーションの WHERE status='published' AND id <= ? ORDER BY id DESC を支える
-- ② GIN(required_skills): スキル検索の @>（配列包含）を支える（BTree では配列の中身を引けない）

-- +migrate Up
CREATE INDEX idx_projects_status_id ON projects (status, id DESC);

CREATE INDEX idx_projects_required_skills ON projects USING gin (required_skills);

-- +migrate Down
DROP INDEX idx_projects_required_skills;

DROP INDEX idx_projects_status_id;
