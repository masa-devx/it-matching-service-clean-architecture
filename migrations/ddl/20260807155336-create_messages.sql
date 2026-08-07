
-- +migrate Up
-- 契約の当事者同士のメッセージ。
--
-- このサービスの生命線である「連絡先マスキング」を担うテーブル。
-- プラットフォーム外で直接契約されると、手数料が入らないだけでなく
-- トラブル時の保護（エスクロー・稼働報告・レビュー）がすべて効かなくなる。
CREATE TABLE messages (
    id          BIGSERIAL PRIMARY KEY,
    -- メッセージは契約に紐づく（合意後のやり取りが対象）。
    -- 応募段階のやり取りは対象外——マスキングが最も重要なのは「合意直前〜直後」のため
    contract_id BIGINT NOT NULL REFERENCES contracts (id) ON DELETE CASCADE,
    -- 送信者は契約の当事者のどちらか。users を参照せずロールで持つのは、
    -- 契約が company_id / talent_id を持っており、ロールが決まれば送信者も一意に決まるため
    sender_role TEXT   NOT NULL CHECK (sender_role IN ('company', 'talent')),

    -- 原文と表示用を分けて保存する。
    --
    -- body: ユーザーが実際に入力した文字列。規約違反の調査や紛争時の証跡として必要なので
    --       消さずに残す。ただし API からは返さない（返さなければ漏れない）。
    -- masked_body: 連絡先を伏せた表示用の文字列。保存時に一度だけ生成する
    --       （表示のたび計算すると、正規表現を改善したときに過去の表示が変わってしまう。
    --        「あのとき相手に何が見えていたか」を再現できる状態を保つ）
    body        TEXT   NOT NULL,
    masked_body TEXT   NOT NULL,
    -- 伏せ字が入ったかどうか。UIで「一部を伏せました」と理由を伝えるために使う
    -- （伏せ字だけ見せられても、ユーザーは何が起きたのか分からない）
    masked      BOOLEAN NOT NULL DEFAULT false,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 「この契約の会話を古い順に」引くための索引。会話は上から下へ読むので昇順
CREATE INDEX messages_contract_created_at_idx ON messages (contract_id, created_at);

COMMENT ON COLUMN messages.sender_role IS '送信者（company / talent）。契約が当事者を持つため、ロールが決まれば送信者が一意に定まる';
COMMENT ON COLUMN messages.body IS '原文。監査・紛争時の証跡として保存するが、APIからは返さない';
COMMENT ON COLUMN messages.masked_body IS '連絡先を伏せた表示用の文字列。保存時に一度だけ生成する';
COMMENT ON COLUMN messages.masked IS '伏せ字が入ったか。UIで理由を伝えるために使う';

-- +migrate Down
DROP TABLE messages;
