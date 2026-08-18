-- 開発用シードデータ。sql-migrate の管理外（本番に混入させない）・`make seed` で投入する。
-- 冪等性の方針:
--   users / プロフィール = upsert（何度実行しても既知の状態に収束）
--   projects            = delete-then-insert（'デモ案件 #' 接頭辞で識別。手動作成データは消さない）
-- ばらつきは連番からの決定的な導出で作る（random() 禁止 = 誰がいつ実行しても同じデータ）。
-- パスワードは全員 password123（開発専用の既知の値。実在の資格情報は絶対に置かない）

BEGIN;

-- ユーザー4人（company 2 / talent 2）
INSERT INTO users (email, password_hash, role) VALUES
    ('company1@example.com', '$2a$10$.YKW8WknefpK6PvJEJJ7jujUN3lkYf9WVIBO72gd3/ZrrNzn7GqIq', 'company'),
    ('company2@example.com', '$2a$10$.YKW8WknefpK6PvJEJJ7jujUN3lkYf9WVIBO72gd3/ZrrNzn7GqIq', 'company'),
    ('talent1@example.com', '$2a$10$.YKW8WknefpK6PvJEJJ7jujUN3lkYf9WVIBO72gd3/ZrrNzn7GqIq', 'talent'),
    ('talent2@example.com', '$2a$10$.YKW8WknefpK6PvJEJJ7jujUN3lkYf9WVIBO72gd3/ZrrNzn7GqIq', 'talent')
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash;

-- 企業プロフィール
INSERT INTO companies (user_id, name, location, description)
SELECT u.id, v.name, v.location, v.description
FROM (VALUES
    ('company1@example.com', '株式会社テックワークス', '東京', '自社SaaSの開発・運営を行っています。'),
    ('company2@example.com', '合同会社リモートラボ', '大阪', 'フルリモートの受託開発チームです。')
) AS v (email, name, location, description)
JOIN users u ON u.email = v.email
ON CONFLICT (user_id) DO UPDATE SET
    name = EXCLUDED.name,
    location = EXCLUDED.location,
    description = EXCLUDED.description;

-- 人材プロフィール
INSERT INTO talents (user_id, display_name, skills, bio)
SELECT u.id, v.display_name, v.skills, v.bio
FROM (VALUES
    ('talent1@example.com', '山田太郎', ARRAY['Go', 'PostgreSQL', 'TypeScript'], 'バックエンド中心に5年。副業で週10〜20時間稼働できます。'),
    ('talent2@example.com', '鈴木花子', ARRAY['React', 'Next.js'], 'フロントエンド専門。UI実装とアクセシビリティが得意です。')
) AS v (email, display_name, skills, bio)
JOIN users u ON u.email = v.email
ON CONFLICT (user_id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    skills = EXCLUDED.skills,
    bio = EXCLUDED.bio;

-- デモ案件 47件（published 42 = 既定 limit 20 でちょうど3ページ / draft 3 / closed 2）
DELETE FROM projects WHERE title LIKE 'デモ案件 #%';

INSERT INTO projects (
    company_id,
    title,
    description,
    hourly_rate_min,
    hourly_rate_max,
    hours_per_week,
    remote_ok,
    required_skills,
    status
)
SELECT
    CASE WHEN gs.i <= 30
        THEN (SELECT c.id FROM companies c JOIN users u ON u.id = c.user_id WHERE u.email = 'company1@example.com')
        ELSE (SELECT c.id FROM companies c JOIN users u ON u.id = c.user_id WHERE u.email = 'company2@example.com')
    END,
    'デモ案件 #' || gs.i,
    'シード生成のデモ案件です（連番 ' || gs.i || ' から決定的に導出）。検索・ページネーションの動作確認に使います。',
    -- 10の倍数は時給未設定（NULL 許容列の表示・絞り込みの確認用）
    CASE WHEN gs.i % 10 = 0 THEN NULL ELSE 3000 + (gs.i % 7) * 500 END,
    CASE WHEN gs.i % 10 = 0 THEN NULL ELSE 4000 + (gs.i % 7) * 500 + (gs.i % 3) * 500 END,
    10 + (gs.i % 21),
    gs.i % 2 = 0,
    sk.skills,
    CASE
        WHEN gs.i IN (7, 21, 35) THEN 'draft'
        WHEN gs.i IN (14, 28) THEN 'closed'
        ELSE 'published'
    END
FROM generate_series(1, 47) AS gs (i)
-- スキルは7種から2〜3個。3も5も7と互いに素なので j=0..2 で必ず異なる添字になる
CROSS JOIN LATERAL (
    SELECT ARRAY(
        SELECT (ARRAY['Go', 'PostgreSQL', 'TypeScript', 'React', 'Next.js', 'AWS', 'Docker'])[((gs.i * 3 + j * 5) % 7) + 1]
        FROM generate_series(0, 1 + (gs.i % 2)) AS j
    ) AS skills
) sk;

COMMIT;

-- 統計情報を更新（大量投入直後はプランナが古い統計で誤った実行計画を選ぶ）
ANALYZE users, companies, talents, projects;
