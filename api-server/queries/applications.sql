-- 応募のクエリ。status は列挙しない（DB の DEFAULT 'applied' に任せる＝応募は必ず applied から始まる）

-- name: CreateApplication :one
-- 「公開中の案件にのみ応募できる」を INSERT...SELECT で原子的に埋め込む。
-- SELECT で公開確認してから INSERT すると、その間に非公開化される窓が空く（条件付きUPDATEと同型の問題）。
-- 0行 = 未公開または不存在（区別せず404）。二重応募は UNIQUE(project_id, talent_id) 違反（23505 → 409）
INSERT INTO applications (project_id, talent_id, message)
SELECT p.id, @talent_id, @message
FROM projects p
WHERE p.id = @project_id AND p.status = 'published'
RETURNING
    id,
    project_id,
    talent_id,
    status,
    message,
    company_acted_at,
    talent_acted_at,
    created_at;

-- name: ListApplicationsForTalent :many
-- 自分の応募一覧（新しい順）。案件タイトルは JOIN で供給する（一覧表示用）。
-- WHERE talent_id ＝ 他人の応募は一覧に存在しない
SELECT
    a.id,
    a.project_id,
    p.title AS project_title,
    a.status,
    a.message,
    a.offer_message,
    a.created_at
FROM applications a
JOIN projects p ON p.id = a.project_id
WHERE a.talent_id = $1
ORDER BY a.id DESC;

-- name: GetApplicationForTalent :one
-- 所有チェック込みの取得（レスポンス組み立て用。project_title 込み）
SELECT
    a.id,
    a.project_id,
    p.title AS project_title,
    a.status,
    a.message,
    a.offer_message,
    a.created_at
FROM applications a
JOIN projects p ON p.id = a.project_id
WHERE a.id = $1 AND a.talent_id = $2;

-- name: ListApplicationsForProject :many
-- 自社案件の応募一覧（company 視点・応募者プロフィール込み・新しい順）。
-- 所有確認（他社案件の404）は呼び出し側が GetProjectForCompany で先に行う:
-- 「他社の案件（404）」と「自社の案件で応募ゼロ（空リスト）」を区別するため、
-- WHERE への埋め込みではなく2クエリに分ける（読み取りなので競合窓の心配はない）
SELECT
    a.id,
    a.project_id,
    a.status,
    a.message,
    a.offer_message,
    t.display_name AS talent_display_name,
    t.skills AS talent_skills,
    a.created_at
FROM applications a
JOIN talents t ON t.id = a.talent_id
WHERE a.project_id = $1
ORDER BY a.id DESC;

-- name: GetApplicationForCompany :one
-- 所有チェックが JOIN 越しになる: applications は company_id を持たないため、
-- projects を結合して p.company_id を WHERE に埋め込む（他社の応募は「存在しない」扱い）
SELECT
    a.id,
    a.project_id,
    a.status,
    a.message,
    a.offer_message,
    t.display_name AS talent_display_name,
    t.skills AS talent_skills,
    a.created_at
FROM applications a
JOIN projects p ON p.id = a.project_id
JOIN talents t ON t.id = a.talent_id
WHERE a.id = $1 AND p.company_id = $2;

-- name: UpdateApplicationStatusForCompany :one
-- 選考の遷移（offer / reject は同じ形なので to_status で共通化・#42 の型）。
-- 所有（JOIN 越しの p.company_id）と遷移可否（from_statuses・遷移表から導出）を
-- DB が原子的に検査する。company_acted_at で「企業がいつ動いたか」も同時に記録
UPDATE applications a
SET
    status = @to_status,
    offer_message = @offer_message,
    company_acted_at = now()
FROM projects p, talents t
WHERE a.id = @id
    AND p.id = a.project_id
    AND p.company_id = @company_id
    AND t.id = a.talent_id
    AND a.status = ANY (@from_statuses::text [])
RETURNING
    a.id,
    a.project_id,
    a.status,
    a.message,
    a.offer_message,
    t.display_name AS talent_display_name,
    t.skills AS talent_skills,
    a.company_acted_at,
    a.created_at;

-- name: UpdateApplicationStatusForTalent :one
-- talent の遷移（withdraw / accept / decline は同じ形なので to_status で共通化・company 側と対称）。
-- 遷移元のホワイトリストは呼び出し側が shared/domain の遷移表から導出して渡す
-- （表が一次情報のまま WHERE に反映される）。
-- WHERE に talent_id と from_statuses を含めることで、所有と遷移可否を DB が原子的に検査する。
-- FROM projects はレスポンス用の project_title を1文で取るための JOIN（UPDATE...FROM）
UPDATE applications a
SET
    status = @to_status,
    talent_acted_at = now()
FROM projects p
WHERE a.id = @id
    AND a.talent_id = @talent_id
    AND p.id = a.project_id
    AND a.status = ANY (@from_statuses::text [])
RETURNING
    a.id,
    a.project_id,
    p.title AS project_title,
    a.status,
    a.message,
    a.offer_message,
    a.talent_acted_at,
    a.created_at;
