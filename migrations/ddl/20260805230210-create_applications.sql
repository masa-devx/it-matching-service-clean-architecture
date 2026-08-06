
-- +migrate Up
-- 応募（人材 → 案件）。状態機械①の本体。
-- projects と 1:N、talents と 1:N の交差テーブルにあたる
CREATE TABLE applications (
    id         BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    -- users ではなく talents を参照する（projects.company_id が companies を指すのと対称）。
    -- 結果として「人材プロフィール未登録では応募できない」という仕様がスキーマで表現される
    talent_id  BIGINT NOT NULL REFERENCES talents (id) ON DELETE CASCADE,
    -- 選考の状態。offered 以降の遷移は #84 で実装するが、
    -- 後からマイグレーションを足さずに済むよう取りうる値をすべて先に定義しておく
    status     TEXT   NOT NULL DEFAULT 'applied'
               CHECK (status IN ('applied', 'offered', 'accepted', 'rejected', 'withdrawn', 'declined')),
    message    TEXT   NOT NULL,
    -- ダブルオプトインの2つの意思表示。「まだ意思表示していない」を表すゼロ値が
    -- 時刻には存在しないため、この2列に限り NULL を許容する（他列の空値方針の例外）
    company_acted_at TIMESTAMPTZ,
    talent_acted_at  TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- 二重応募をDBで物理的に拒否する。アプリ側の存在確認だけでは
    -- 確認と挿入の間に別リクエストが割り込むと通ってしまう（TOCTOU）
    CONSTRAINT applications_project_talent_key UNIQUE (project_id, talent_id)
);

-- talent_id 単独の索引。上の UNIQUE は (project_id, talent_id) の複合索引を兼ねるため
-- 「この案件への応募一覧」には効くが、先頭列を使わない「自分の応募一覧」には効かない
CREATE INDEX applications_talent_id_idx ON applications (talent_id);

-- 「この案件の応募を状態で絞って引く」ための索引（企業の応募者管理画面）
CREATE INDEX applications_project_status_idx ON applications (project_id, status);

COMMENT ON COLUMN applications.project_id IS '応募先の案件（projects.id）';
COMMENT ON COLUMN applications.talent_id IS '応募した人材（talents.id）。users.id ではないので注意';
COMMENT ON COLUMN applications.status IS '選考状態: applied=応募済み / offered=企業がオファー / accepted=成立 / rejected=企業が見送り / withdrawn=人材が取り下げ / declined=人材が辞退';
COMMENT ON COLUMN applications.message IS '志望動機。企業が選考するための本文';
COMMENT ON COLUMN applications.company_acted_at IS '企業が意思表示した時刻（オファー・見送り）。NULL=未実施';
COMMENT ON COLUMN applications.talent_acted_at IS '人材が意思表示した時刻（承諾・辞退・取り下げ）。NULL=未実施';

-- +migrate Down
DROP TABLE applications;
