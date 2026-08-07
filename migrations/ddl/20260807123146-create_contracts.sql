
-- +migrate Up
-- 契約（応募の承諾で成立する取引の単位）。状態機械②の本体。
-- ダブルオプトインが成立した瞬間に applications から自動生成される
CREATE TABLE contracts (
    id             BIGSERIAL PRIMARY KEY,
    -- 1応募につき契約は1つ。UNIQUE で 1:1 をDBが保証する
    -- （アプリ側のチェックだけでは同時リクエストで二重に作られうる）
    application_id BIGINT NOT NULL UNIQUE REFERENCES applications (id) ON DELETE CASCADE,
    -- どの案件から生まれたかを辿るための参照。条件は下の列にコピー済みなので、
    -- 表示・監査のためだけに持つ（契約の内容はこの参照に依存しない）
    project_id     BIGINT NOT NULL REFERENCES projects (id),
    -- 契約の当事者。application → project → company と辿れるが、
    -- 契約一覧で毎回3テーブルを結合するのは重いため、当事者は直接持つ
    company_id     BIGINT NOT NULL REFERENCES companies (id) ON DELETE CASCADE,
    talent_id      BIGINT NOT NULL REFERENCES talents (id) ON DELETE CASCADE,

    -- ここから承諾時点の条件のスナップショット。
    -- 案件は掲載後に編集できる（PUT /projects/{id}）ため、参照のままだと
    -- 契約成立後に単価を書き換えられてしまう。合意した内容を契約側で固定する
    title          TEXT    NOT NULL,
    hourly_rate    INT     NOT NULL,
    hours_per_week INT     NOT NULL,
    remote_ok      BOOLEAN NOT NULL,

    -- 契約の進行。遷移は #104 の遷移表で一元管理するが、
    -- 後からマイグレーションを足さずに済むよう取りうる値をすべて定義しておく
    status         TEXT    NOT NULL DEFAULT 'active'
                   CHECK (status IN ('active', 'working', 'reviewing', 'completed', 'cancelled')),
    -- 「まだ開始していない」を表すゼロ値が時刻には無いため NULL を許容する
    started_at     TIMESTAMPTZ,
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- 単価は募集レンジの上限を採用するため、0以上であることだけ保証する
    CONSTRAINT contracts_hourly_rate_check CHECK (hourly_rate >= 0),
    CONSTRAINT contracts_hours_check CHECK (hours_per_week >= 0 AND hours_per_week <= 168)
);

-- 「自社（自分）の契約を新しい順に」引くための索引。当事者ごとに1本ずつ張る
CREATE INDEX contracts_company_id_created_at_idx ON contracts (company_id, created_at DESC);
CREATE INDEX contracts_talent_id_created_at_idx  ON contracts (talent_id, created_at DESC);

-- 状態での絞り込み（稼働中だけ、検収待ちだけ）
CREATE INDEX contracts_status_idx ON contracts (status);

COMMENT ON COLUMN contracts.application_id IS '成立した応募（applications.id）。UNIQUE で1応募1契約を保証';
COMMENT ON COLUMN contracts.project_id IS '契約のもとになった案件。条件は本テーブルにコピー済みで、表示・監査のための参照';
COMMENT ON COLUMN contracts.title IS '承諾時点の案件名のコピー。案件が改題されても契約の記載は変わらない';
COMMENT ON COLUMN contracts.hourly_rate IS '合意した時給・円。募集レンジの上限を採用する（条件交渉はM4以降）';
COMMENT ON COLUMN contracts.status IS '契約の進行: active=成立 / working=稼働中 / reviewing=検収待ち / completed=完了 / cancelled=中止';
COMMENT ON COLUMN contracts.started_at IS '稼働を開始した時刻。NULL=未開始';
COMMENT ON COLUMN contracts.completed_at IS '検収が完了した時刻。NULL=未完了';

-- +migrate Down
DROP TABLE contracts;
