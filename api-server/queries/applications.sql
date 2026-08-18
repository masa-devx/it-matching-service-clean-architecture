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
    a.created_at
FROM applications a
JOIN projects p ON p.id = a.project_id
WHERE a.id = $1 AND a.talent_id = $2;

-- name: WithdrawApplication :one
-- 取り下げ（talent の遷移）。遷移元のホワイトリストは呼び出し側が
-- shared/domain の遷移表から導出して渡す（表が一次情報のまま WHERE に反映される）。
-- WHERE に talent_id と from_statuses を含めることで、所有と遷移可否を DB が原子的に検査する
UPDATE applications
SET
    status = 'withdrawn',
    talent_acted_at = now()
WHERE id = @id
    AND talent_id = @talent_id
    AND status = ANY (@from_statuses::text [])
RETURNING
    id,
    project_id,
    talent_id,
    status,
    message,
    company_acted_at,
    talent_acted_at,
    created_at;
