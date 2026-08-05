-- 性能計測用クエリ集（実行: make perf）
--
-- API が実際に発行するSQLと同じ形にしてある。誰が実行しても同じ手順・同じ対象を
-- 測れるようにファイル化しておくことで、before/after の比較が成立する。
--
-- 読み方:
--   Seq Scan            = 全行を順に読む（大きなテーブルでは赤信号）
--   Index Scan          = 索引を辿って必要な行だけ読む
--   Bitmap Index Scan   = 索引で行の位置を集めてからまとめて読む（中量ヒット時に選ばれる）
--   cost=..             = プランナの見積もり（相対値。単位は「ページ読み取り1回=1.0」）
--   actual time=..      = 実測ミリ秒（開始..終了）
--   rows=..             = 見積もりと実測の行数。大きくズレていたら ANALYZE 不足を疑う
--   Buffers: shared hit = メモリから読めたブロック数 / read = ディスクから読んだ数

\echo '===== [1] 一覧 1ページ目（公開中を新しい順・LIMIT 20）====='
EXPLAIN (ANALYZE, BUFFERS)
SELECT p.id, p.company_id, c.name, p.title, p.required_skills, p.created_at
FROM projects p
JOIN companies c ON c.id = p.company_id
WHERE p.status = 'published'
ORDER BY p.created_at DESC, p.id DESC
LIMIT 20 OFFSET 0;

\echo '===== [2] 深いページ（OFFSET 40000）— OFFSET の劣化を見る ====='
EXPLAIN (ANALYZE, BUFFERS)
SELECT p.id, p.company_id, c.name, p.title, p.required_skills, p.created_at
FROM projects p
JOIN companies c ON c.id = p.company_id
WHERE p.status = 'published'
ORDER BY p.created_at DESC, p.id DESC
LIMIT 20 OFFSET 40000;

\echo '===== [3] スキル検索（GIN インデックスが効くか）====='
EXPLAIN (ANALYZE, BUFFERS)
SELECT p.id, p.title, p.required_skills
FROM projects p
JOIN companies c ON c.id = p.company_id
WHERE p.status = 'published' AND p.required_skills @> ARRAY['Go']
ORDER BY p.created_at DESC, p.id DESC
LIMIT 20;

\echo '===== [4] 複合条件（スキル＋単価＋リモート＋稼働）====='
EXPLAIN (ANALYZE, BUFFERS)
SELECT p.id, p.title, p.hourly_rate_min, p.hourly_rate_max
FROM projects p
JOIN companies c ON c.id = p.company_id
WHERE p.status = 'published'
  AND p.required_skills @> ARRAY['Go']
  AND p.hourly_rate_max >= 8000
  AND p.hours_per_week <= 20
  AND p.remote_ok = true
ORDER BY p.created_at DESC, p.id DESC
LIMIT 20;

\echo '===== [5] 総件数（ページャ表示用）— 全件走査になるか ====='
EXPLAIN (ANALYZE, BUFFERS)
SELECT count(*) FROM projects p WHERE p.status = 'published';

\echo '===== [6] キーワード部分一致（ILIKE %...% — B-tree が効かない形）====='
EXPLAIN (ANALYZE, BUFFERS)
SELECT p.id, p.title
FROM projects p
JOIN companies c ON c.id = p.company_id
WHERE p.status = 'published' AND p.title ILIKE '%00042%'
ORDER BY p.created_at DESC, p.id DESC
LIMIT 20;

\echo '===== 参考: テーブルとインデックスのサイズ ====='
SELECT
    relname AS object,
    pg_size_pretty(pg_relation_size(oid)) AS size
FROM pg_class
WHERE relname LIKE 'projects%'
ORDER BY pg_relation_size(oid) DESC;
