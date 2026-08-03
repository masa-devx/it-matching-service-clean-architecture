# 学習ログ: DB

> 各ステップの型: **何をした / 概念 / ✅ベストプラクティス / ⚠️アンチパターン / 理解度チェック**
> 目次: [学習ログ.md](学習ログ.md)

---

## Phase0 #3 Step1: users テーブルと sql-migrate 導入

**何をした**: sql-migrate（it-support-service と同じ）を導入し、最初のマイグレーション `migrations/ddl/20260803000001-create_users.sql` で users テーブルを定義（email UNIQUE / password_hash / role CHECK / TIMESTAMPTZ）。`dbconfig.yml`・`Makefile`（up/down/status/new）を整備。

**概念**

- **マイグレーション管理**: DDLを日付順のファイルで積み上げ、適用履歴を `gorp_migrations` テーブルに記録。「どこまで適用済みか」をDB自身が知っているので、同じファイルが二重適用されない（＝`IF NOT EXISTS` に頼る必要がなくなる）
- **Up / Down**: Up=適用、Down=巻き戻し。Downを毎回書く規律が「変更を戻せる設計」を強制する。ただし本番でのdownは原則禁止（DROP TABLEでデータが消える。戻したいときは「逆操作のUp」を新規作成する＝ロールフォワード）
- `BIGSERIAL`: 自動採番の64bit整数（`SERIAL`=32bitは上限約21億。ユーザー系テーブルは最初からBIGで困らない）
- `UNIQUE` 制約: DBが重複を物理的に拒否する最後の砦。アプリ側チェックだけでは同時リクエストの競合（TOCTOU）を防げない
- `CHECK (role IN (...))`: 列挙値の制約。PostgreSQL の `ENUM` 型と違い、値の追加が制約の張り替えで済み、マイグレーションが軽い
- `TIMESTAMPTZ`: タイムゾーン付き時刻。`TIMESTAMP`（tzなし）は「どの地域の時刻か」が失われ事故のもと

**✅ベストプラクティス**: スキーマ変更は必ずマイグレーションファイル経由（psqlで直接DDLを流さない）／パスワードは `password_hash` 列名にして「ハッシュしか入らない」ことを名前で強制する／時刻列は `TIMESTAMPTZ NOT NULL DEFAULT now()`

**⚠️アンチパターン**: 適用済みマイグレーションファイルを後から編集する（履歴とDBの実態がズレる。修正は新しいファイルで）／本番での安易な down／`TIMESTAMP`（tzなし）／アプリ側の重複チェックだけに頼って UNIQUE を張らない

**理解度チェック**: gorp_migrations テーブルの役割は？ 本番で down を使わずロールフォワードする理由は？ UNIQUE制約がアプリ側チェックより強い理由を同時リクエストの観点で説明できるか？
