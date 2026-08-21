-- 応募へのオファーメッセージ（#98）。
-- NULL 許容で追加する: 既存行に影響を与えず（バックフィル・ロック不要）、
-- 「オファー時にメッセージが無かった」を NULL で自然に表現できる。
-- 長さの上限（500文字）は DB の CHECK ではなく validator（Go）が唯一の防衛線（既存の方針どおり）

-- +migrate Up
ALTER TABLE applications ADD COLUMN offer_message TEXT;

-- +migrate Down
ALTER TABLE applications DROP COLUMN offer_message;
