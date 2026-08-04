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
