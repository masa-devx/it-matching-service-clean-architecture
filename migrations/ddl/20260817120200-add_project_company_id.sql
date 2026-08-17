-- 案件に所有者（企業）を付ける。まず NULL 許容で列を追加し、
-- 認証が繋がって全行に所有者が付く #31 で NOT NULL 化する（2段階の後方互換な変更）。
-- PostgreSQL は FK に索引を自動作成しないため明示的に張る（JOIN・CASCADE の性能）

-- +migrate Up
ALTER TABLE projects
    ADD COLUMN company_id BIGINT REFERENCES companies (id) ON DELETE CASCADE;

CREATE INDEX idx_projects_company_id ON projects (company_id);

-- +migrate Down
DROP INDEX idx_projects_company_id;

ALTER TABLE projects
    DROP COLUMN company_id;
