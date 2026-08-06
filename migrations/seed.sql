-- 開発用シードデータ（ローカル専用・本番では実行しない）
-- 実行: make seed ／ 2回実行しても安全（upsert）
--
-- ここはマイグレーション（ddl/）ではない: スキーマは全環境で適用するが、
-- テストユーザーはローカル開発にだけ必要なため、sql-migrate の管理外に置く

-- パスワードは全員 'password123'（README参照）。
-- ハッシュは既知のdev用パスワードを事前計算したもので、秘密情報ではない
INSERT INTO users (email, password_hash, role)
VALUES
    ('company1@example.com', '$2a$10$FLINvcBjmj.iREn0o5hZfeoSMa5FHtRFWW0b2SwhGfjCBEfWi5JQa', 'company'),
    ('company2@example.com', '$2a$10$FLINvcBjmj.iREn0o5hZfeoSMa5FHtRFWW0b2SwhGfjCBEfWi5JQa', 'company'),
    ('talent1@example.com',  '$2a$10$FLINvcBjmj.iREn0o5hZfeoSMa5FHtRFWW0b2SwhGfjCBEfWi5JQa', 'talent'),
    ('talent2@example.com',  '$2a$10$FLINvcBjmj.iREn0o5hZfeoSMa5FHtRFWW0b2SwhGfjCBEfWi5JQa', 'talent')
-- DO NOTHING ではなく DO UPDATE: 過去の手動テストで同じemailを別パスワードで
-- 作っていても、シード実行後は必ず既知の資格情報でログインできる状態に収束させる
ON CONFLICT (email) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    role          = EXCLUDED.role,
    updated_at    = now();

-- プロフィール。案件の掲載には companies、応募には talents が必要なため、
-- ログイン直後から一通りの動作確認ができるよう最初から入れておく。
-- user_id は email から引く（連番のIDに依存させない＝実行順序が変わっても壊れない）
INSERT INTO companies (user_id, name, description, industry, size)
SELECT id, v.name, v.description, v.industry, v.size
FROM users
JOIN (VALUES
    ('company1@example.com', '株式会社ツナグ', '受託開発とSaaSの2軸で事業を展開しています。', 'IT・ソフトウェア', '11-50'),
    ('company2@example.com', 'ワークス合同会社', '業務システムの内製化を支援しています。', 'IT・ソフトウェア', '1-10')
) AS v(email, name, description, industry, size) ON v.email = users.email
ON CONFLICT (user_id) DO UPDATE
SET name        = EXCLUDED.name,
    description = EXCLUDED.description,
    industry    = EXCLUDED.industry,
    size        = EXCLUDED.size,
    updated_at  = now();

INSERT INTO talents (user_id, display_name, bio, skills, years_of_exp, available_hours_per_week, desired_hourly_rate, remote_ok)
SELECT id, v.display_name, v.bio, v.skills, v.years_of_exp, v.hours, v.rate, v.remote_ok
FROM users
JOIN (VALUES
    ('talent1@example.com', '山田 太郎', 'バックエンド中心に受託・自社サービスを経験。', ARRAY['Go', 'PostgreSQL', 'AWS'], 5, 20, 6000, true),
    ('talent2@example.com', '田中 花子', 'フロントエンド専門。デザインシステムの構築が得意。', ARRAY['React', 'TypeScript', 'Next.js'], 3, 10, 5000, true)
) AS v(email, display_name, bio, skills, years_of_exp, hours, rate, remote_ok) ON v.email = users.email
ON CONFLICT (user_id) DO UPDATE
SET display_name             = EXCLUDED.display_name,
    bio                      = EXCLUDED.bio,
    skills                   = EXCLUDED.skills,
    years_of_exp             = EXCLUDED.years_of_exp,
    available_hours_per_week = EXCLUDED.available_hours_per_week,
    desired_hourly_rate      = EXCLUDED.desired_hourly_rate,
    remote_ok                = EXCLUDED.remote_ok,
    updated_at               = now();
