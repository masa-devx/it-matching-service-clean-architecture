-- 性能計測用の大量シード（案件5万件・ローカル専用）
-- 実行: make seed-large ／ 何度実行しても結果は同じ（生成分を消してから作り直す）
--
-- 設計方針:
-- 1. generate_series で一括生成する（1行ずつ INSERT すると数万件で数分かかる）
-- 2. random() を使わず連番 i から決定的に値を導出する
--    → 誰がいつ実行しても同じデータになり、性能計測の before/after を比較できる
-- 3. 生成データは title / email の接頭辞で識別し、手動作成データを壊さない

BEGIN;

-- 生成分だけを削除（projects は companies の CASCADE で消えるが、明示して意図を示す）
DELETE FROM projects WHERE title LIKE 'デモ案件 #%';
DELETE FROM users WHERE email LIKE 'seed-company%@example.com';

-- ① 掲載元の企業を10社（案件を1社に集中させると company_id の索引が意味を持たない）
INSERT INTO users (email, password_hash, role)
SELECT
    'seed-company' || i || '@example.com',
    -- パスワードは password123（seed.sql と同じ既知のハッシュ）
    '$2a$10$FLINvcBjmj.iREn0o5hZfeoSMa5FHtRFWW0b2SwhGfjCBEfWi5JQa',
    'company'
FROM generate_series(1, 10) AS i;

INSERT INTO companies (user_id, name, description, industry, size)
SELECT
    u.id,
    'シード企業' || i,
    'パフォーマンス計測用のダミー企業です',
    (ARRAY['SaaS', '受託開発', 'EC', '金融', 'メディア'])[1 + (i % 5)],
    (ARRAY['1-10', '11-50', '51-200', '201+'])[1 + (i % 4)]
FROM generate_series(1, 10) AS i
JOIN users u ON u.email = 'seed-company' || i || '@example.com';

-- ② 案件を5万件
INSERT INTO projects (
    company_id, title, description, required_skills,
    hourly_rate_min, hourly_rate_max, hours_per_week, remote_ok, status, created_at, updated_at
)
SELECT
    c.id,
    -- 連番を0埋めして通常シード・手動作成と区別できるようにする
    'デモ案件 #' || lpad(i::text, 5, '0'),
    'これは性能計測用に生成されたダミー案件です（連番 ' || i || '）',
    -- スキルは14種のプールから 1〜4 個を選ぶ。
    -- 添字は「i で開始位置をずらし、j で 5 ずつ進める」= (i*3 + j*5) % 14。
    -- 3 も 5 も 14 と互いに素なので、j=0..3 で必ず異なる4つの添字になる
    -- （乗数が 14 と公約数を持つ / 剰余が一致する組み合わせだと同じスキルに偏る）
    (
        SELECT array_agg(skills.name)
        FROM (
            SELECT (ARRAY[
                'Go', 'TypeScript', 'React', 'Next.js', 'PostgreSQL', 'AWS', 'Docker',
                'Kubernetes', 'Python', 'Ruby', 'Rails', 'Vue.js', 'GraphQL', 'Terraform'
            ])[1 + ((i * 3 + j * 5) % 14)] AS name
            FROM generate_series(0, i % 4) AS j  -- 個数も i によって 1〜4 個に変わる
        ) AS skills
    ),
    rate.min_rate,
    -- 上限は下限 + 1000〜3000（CHECK制約 min <= max を必ず満たす）
    rate.min_rate + 1000 + ((i * 11) % 3) * 1000,
    5 + ((i * 17) % 36),                          -- 稼働 5〜40h/週
    (i * 3) % 5 < 3,                              -- 約60%がリモート可
    CASE
        WHEN (i * 19) % 100 < 85 THEN 'published' -- 85%
        WHEN (i * 19) % 100 < 95 THEN 'draft'     -- 10%
        ELSE 'closed'                             -- 5%
    END,
    -- 直近2年間に分散させる（新しい順の並び替え・seekページングの計測に必要）
    now() - ((i * 13) % 730) * interval '1 day' - ((i * 7) % 86400) * interval '1 second',
    now()
FROM generate_series(1, 50000) AS i
-- 10社へ均等に振り分ける
JOIN companies c ON c.name = 'シード企業' || (1 + (i % 10))
CROSS JOIN LATERAL (SELECT 3000 + ((i * 29) % 13) * 500 AS min_rate) AS rate;

COMMIT;

-- 統計情報を更新して、実行計画が実データに基づくようにする
ANALYZE projects;
