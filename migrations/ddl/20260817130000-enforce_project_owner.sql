-- 所有列の NOT NULL 化（#27 で NULL 許容で追加した列の、2段階変更の第2段）。
-- 所有者の無い既存行は Phase 0 の検証データのみのため削除する
-- （本番想定ではバックフィル→NOT NULL 化の順にするが、所有者が決められない行は本質的に不正データ）

-- +migrate Up
DELETE FROM projects WHERE company_id IS NULL;

ALTER TABLE projects
    ALTER COLUMN company_id SET NOT NULL;

-- +migrate Down
ALTER TABLE projects
    ALTER COLUMN company_id DROP NOT NULL;
