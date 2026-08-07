
-- +migrate Up
-- 相互レビュー。両者が提出するまで非公開にし、揃った時点で同時公開する。
--
-- 【なぜ同時公開か】
-- 提出したら即公開する方式だと、後から書く側が相手の評価を見られてしまう。
-- すると「★5をもらったからこちらも★5にしよう」（忖度）や
-- 「★2をつけられた、仕返しに★1」（報復）が起きて、評価が実態を表さなくなる。
-- 両者が提出するまで互いに見えなければ、正直な評価が集まる（Airbnb 方式）。
CREATE TABLE reviews (
    id          BIGSERIAL PRIMARY KEY,
    contract_id BIGINT NOT NULL REFERENCES contracts (id) ON DELETE CASCADE,
    -- 誰が書いたレビューか。相手が誰かは契約から辿れるため、評価対象は列に持たない
    reviewer_role TEXT NOT NULL CHECK (reviewer_role IN ('company', 'talent')),

    -- 総合評価のみ。技術力・コミュニケーション・納期のような複数軸にすると
    -- 入力の手間が増えて提出率が下がる。両者提出が公開条件である以上、
    -- 提出率は機能が成立するかどうかを決める
    rating  INT  NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT NOT NULL,

    -- 【この2列が同時公開の実装そのもの】
    -- submitted_at: 書いた本人が提出した時刻。提出＝確定で、あとから編集できない
    --               （公開後に書き換えられると同時公開の意味が失われる）
    -- published_at: 双方に見えるようになった時刻。NULL = まだ相手が提出していない。
    --               2人目が提出した瞬間に、両方の行へ同じ時刻をセットする
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- 1契約につき各当事者1件まで。二重投稿をDBで物理的に拒否する
    -- （アプリ側の存在確認だけでは、同時リクエストで2件入りうる）
    CONSTRAINT reviews_contract_reviewer_key UNIQUE (contract_id, reviewer_role)
);

-- 「この契約のレビュー」を引くための索引。上の UNIQUE が
-- (contract_id, reviewer_role) の複合索引を兼ねるため、追加の索引は不要
-- （contract_id が先頭列なので、契約単位の取得に効く）

COMMENT ON COLUMN reviews.reviewer_role IS 'レビューを書いた側（company / talent）。評価対象は契約から辿れるため列に持たない';
COMMENT ON COLUMN reviews.rating IS '総合評価 1〜5。複数軸にしないのは提出率を下げないため';
COMMENT ON COLUMN reviews.submitted_at IS '提出した時刻。提出＝確定で編集不可';
COMMENT ON COLUMN reviews.published_at IS '公開された時刻。NULL=相手が未提出のため非公開。両者が揃った瞬間に両方へ同じ時刻が入る';

-- +migrate Down
DROP TABLE reviews;
