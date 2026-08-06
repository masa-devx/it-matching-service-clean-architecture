
-- +migrate Up
-- 企業が応募者を識別するための表示名。companies.name と対になる列。
-- 既存行があるため DEFAULT '' が必要（NOT NULL 列を後から足すときの定石）。
-- 「必須なのに空を許す」ギャップはアプリ側のバリデーションで埋め、
-- 既存ユーザーは次回のプロフィール保存時に入力する（段階的な必須化）
ALTER TABLE talents ADD COLUMN display_name TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN talents.display_name IS '企業に見せる表示名。メール等の連絡先は別途マスキング対象のため、ここには本名/通称のみを入れる';

-- +migrate Down
ALTER TABLE talents DROP COLUMN display_name;
