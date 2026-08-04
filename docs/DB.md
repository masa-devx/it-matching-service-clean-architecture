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

---

## Phase0 #31: シードデータ（upsert・冪等な投入）

### 何をした
`migrations/seed.sql`（開発用ユーザー4人・パスワードは既知の password123）を作成し、`make seed` で投入。`ON CONFLICT (email) DO UPDATE` で何度実行しても既知の状態に収束する。

### 概念
- **シード ≠ マイグレーション**: スキーマ（ddl/）は全環境で適用するが、テストユーザーはローカル専用。sql-migrate の管理外（migrations/seed.sql）に置き、本番に混入しない構造にする
- **ON CONFLICT（upsert）の2方式**:
  - `DO NOTHING`: 既存行を尊重（重複を静かにスキップ）。「初回だけ入れたい」データ向け
  - `DO UPDATE`: 既存行を上書きし**必ず既知の状態に収束**させる。「実行後の状態を保証したい」シード向け（過去の手動テストで同emailが別パスワードで存在してもログイン可能を保証）
- **EXCLUDED 疑似テーブル**: 衝突時に「挿入しようとした行」の値を参照する構文（`SET password_hash = EXCLUDED.password_hash`）
- **`psql -v ON_ERROR_STOP=1`**: エラー時に即終了させる（既定では続行してしまい、失敗に気づけない）
- 事前計算した bcrypt ハッシュのコミットは「既知のdev用パスワード」だから許される（実在の資格情報は絶対に置かない）

### ✅ベストプラクティス
- シードは冪等に（何度叩いても同じ結果）
- テストユーザーの資格情報はREADMEに明記（探させない）

### ⚠️アンチパターン
- シードをマイグレーション（ddl/）に入れて本番にテストユーザーを撒く
- シードのたびに手でユーザー登録（手順が属人化・状態が不定に）

### 理解度チェック
- [ ] DO NOTHING と DO UPDATE の使い分け基準を言える
- [ ] EXCLUDED が何を指すか言える
- [ ] シードを ddl/ に置いてはいけない理由を言える
