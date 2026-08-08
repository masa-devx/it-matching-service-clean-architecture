# ADR-0001: PostgreSQL を維持する

- ステータス: Accepted
- 日付: 2026-08-08

## 背景

後継リポジトリのDB選定にあたり、「MySQL 互換の分散DBも経験しておくか」という観点で TiDB Serverless を検討した。
一方、出発点である tsunagu-works の SQL には PostgreSQL 固有機能が広く使われている:

| PostgreSQL 固有機能 | 使用箇所 |
| --- | --- |
| `RETURNING` | 13箇所 |
| `TEXT[]`（配列型） | 2テーブル |
| GIN 索引 | 2箇所 |
| `count(...) FILTER (WHERE ...)` | 4箇所 |
| `@>`（配列の包含） | 2箇所 |
| `date_trunc` | 2箇所 |

本リポジトリは sqlc を採用するため（[ADR-0002](0002-no-repository-layer.md)）、**手書きSQLがそのまま生成の入力資産になる**。

## 決定

DB は PostgreSQL を維持する。ローカルは Docker、本番は Neon（無料枠）を使う。

## 代替案と却下理由

| 案 | 却下理由 |
| --- | --- |
| TiDB Serverless（MySQL互換） | 上記の固有機能が全て書き直しになる。「MySQL も経験する」ことより **SQL を資産として持ち越す**ほうを取った |
| Cloud SQL for PostgreSQL | 個人開発には過剰（月3,000円前後）。無料枠のある Neon で十分 |

## 影響

- tsunagu-works の SQL を `queries/*.sql`（sqlc の入力）へほぼそのまま流用できる
- MySQL 系の実務経験は本プロジェクトでは得られない（別テーマとして切り離す）
- 見直しの条件: 分散DBが必要な規模・要件が実際に発生したとき（現状の想定にはない）
