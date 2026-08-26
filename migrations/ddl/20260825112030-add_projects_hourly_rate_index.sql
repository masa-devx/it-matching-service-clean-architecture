-- 時給の下限検索（hourly_rate_min >= ?）を支える部分索引（P-3・#116）。
--
-- 根拠（P-2・#115 の計測・100万件）: 該当が少ない/無い時給条件で Parallel Seq Scan に落ち、
-- 30ms・Buffers read 16,187（全行走査）。LIMIT の早期打ち切りは「見つかるまで読む」ため、
-- 該当0件のとき最悪になる。HTTP 層では負荷下で p50 117ms（ヒットありの同エンドポイントは 2.9ms）。
--
-- 部分索引にする理由:
--   status = 'published'        検索対象は公開案件のみ（draft / closed は決して検索されない）
--   hourly_rate_min IS NOT NULL B-tree は NULL も格納するため、明示しないと時給未設定（約4割）も索引に入る。
--                               クエリの hourly_rate_min >= ? は NULL を通さないので、この述語を満たすことを
--                               プランナが証明でき、索引が使われる
-- 書き込み側のコスト（索引の税金）は docs/パフォーマンス.md に計測値を記録する

-- +migrate Up
CREATE INDEX idx_projects_published_hourly_rate_min
    ON projects (hourly_rate_min)
    WHERE status = 'published' AND hourly_rate_min IS NOT NULL;

-- +migrate Down
DROP INDEX idx_projects_published_hourly_rate_min;
