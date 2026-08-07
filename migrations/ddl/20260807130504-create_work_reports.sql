
-- +migrate Up
-- 週次の稼働報告。人材が提出し、企業が承認または差し戻す。
--
-- 「働きぶりが見えない」という企業側の不安と、「作業した証拠を残したい」という
-- 人材側の要求を同時に解決する、このサービスの独自機能のひとつ。
CREATE TABLE work_reports (
    id          BIGSERIAL PRIMARY KEY,
    contract_id BIGINT NOT NULL REFERENCES contracts (id) ON DELETE CASCADE,

    -- 報告対象の週。「第何週か」を文字列（2026-W32 など）で持つと年跨ぎで壊れるため
    -- （12/29〜1/4 は同じ週なのに年が変わる）、週の開始日そのものを DATE で持つ。
    -- 値は必ず月曜に丸めて保存する（アプリ側で date_trunc('week', ...) を通す）。
    -- 日付そのものなので比較も並べ替えも素直にできる
    week_start  DATE   NOT NULL,

    -- その週に稼働した時間。1週間は168時間なので、それを超える申告は入力ミス
    hours       INT    NOT NULL,
    -- 作業内容の要約。何をしたかが分からない報告は検収の材料にならないため必須
    summary     TEXT   NOT NULL,

    -- 提出後の状態。遷移は api/work_report_status.go の遷移表で一元管理する
    status      TEXT   NOT NULL DEFAULT 'submitted'
                CHECK (status IN ('submitted', 'approved', 'rejected')),
    -- 差し戻しの理由。企業が「何を直せばよいか」を伝えるための欄で、
    -- 承認時は空のままになるため NOT NULL DEFAULT '' とする（時刻と違い空値で表現できる）
    review_note TEXT   NOT NULL DEFAULT '',

    -- 提出・確認の時刻。再提出のたび submitted_at は更新する
    -- （「いつ出された報告か」は最新の提出時刻が意味を持つため）
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- 企業が承認・差し戻しをした時刻。NULL＝まだ確認していない
    reviewed_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- 同じ週の報告を二重に提出できないようにする。アプリ側の存在確認だけでは
    -- 確認と挿入の間に別リクエストが割り込むと通ってしまう（TOCTOU）
    CONSTRAINT work_reports_contract_week_key UNIQUE (contract_id, week_start),
    -- 1週間は168時間。ドメイン知識をDBにも書いておくと、どの経路の INSERT でも守られる
    CONSTRAINT work_reports_hours_check CHECK (hours >= 0 AND hours <= 168)
);

-- 「この契約の報告を新しい週から」引くための索引。
-- 上の UNIQUE は (contract_id, week_start) の複合索引を兼ねるが、
-- 並び順が昇順なので、降順で引く一覧のために DESC の索引を別に張る
CREATE INDEX work_reports_contract_week_desc_idx ON work_reports (contract_id, week_start DESC);

-- 「未確認の報告がある契約」を数えるための索引（企業の一覧で使う）
CREATE INDEX work_reports_status_idx ON work_reports (status);

COMMENT ON COLUMN work_reports.week_start IS '報告対象の週の開始日（必ず月曜）。date_trunc(week) で丸めて保存する';
COMMENT ON COLUMN work_reports.hours IS 'その週の稼働時間。0〜168（1週間の総時間）';
COMMENT ON COLUMN work_reports.status IS 'submitted=提出済み（確認待ち） / approved=承認 / rejected=差し戻し';
COMMENT ON COLUMN work_reports.review_note IS '差し戻しの理由。承認時は空文字';
COMMENT ON COLUMN work_reports.submitted_at IS '最後に提出した時刻。再提出のたび更新される';
COMMENT ON COLUMN work_reports.reviewed_at IS '企業が承認・差し戻しをした時刻。NULL=未確認';

-- +migrate Down
DROP TABLE work_reports;
