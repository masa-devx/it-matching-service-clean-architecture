# ADR-0008: 実DBテストの分離は pgx.Tx で行う（go-txdb 不採用）

- ステータス: Accepted
- 日付: 2026-08-17

## 背景

設計プラン§10 では実DBテストの分離に go-txdb を予定していた（テストごとに BEGIN → 終了時に自動 ROLLBACK）。

しかし本リポジトリの sqlc は `sql_package: pgx/v5` で生成しており（#7）、生成された `Queries` は pgx のインターフェース（`DBTX`）に依存する。**go-txdb は `database/sql` のドライバとして実装されている**ため、pgx ネイティブの生成物とは型が接続できない。

一方、go-txdb の本質（テストごとの BEGIN→ROLLBACK）は pgx だけで実現できる。`pgx.Tx` は sqlc の `DBTX` インターフェース（Exec / Query / QueryRow）を満たすため、`db.New(tx)` でトランザクション上の `Queries` がそのまま作れる。

## 決定

テストの分離は **pgx.Tx で実装**する（ライブラリ追加なし）:

```go
tx, _ := conn.Begin(ctx)
t.Cleanup(func() { _ = tx.Rollback(ctx) })  // テスト終了時に自動で元通り
return db.New(tx)                            // pgx.Tx は DBTX を満たす
```

「go-txdb という道具」ではなく「テストごとに BEGIN→ROLLBACK という考え方」を採用の単位とし、テスト戦略（実DB・モックしない・同一インスタンスの別DB）は設計プラン§10 のまま維持する。

## 代替案と却下理由

| 案 | 却下理由 |
| --- | --- |
| go-txdb を使うため sqlc を `database/sql` 生成に変更 | テストの都合で本体の DB アクセス方式を後退させる本末転倒。pgx の性能・型の利点を失う |
| go-txdb ＋ stdlib アダプタ（`*sql.DB`）を併用 | `*sql.DB` は pgx 生成の `Queries` と型が合わず、テスト専用の別 Queries が必要になる（テストと本体で別物を検証することになる） |
| testcontainers 等でテストごとに DB を作る | 分離の粒度がコンテナ単位で遅い。「同一インスタンスの別DB＋Tx分離」で十分 |

## 影響

- ヘルパーは約10行で完結し、依存が増えない（`test/helpers/db.go`）
- テスト対象の usecase 内で明示的なトランザクション（`Queries.WithTx`）を張るようになった場合、テストのTxと入れ子になる。pgx はネストを SAVEPOINT で表現するため基本は動くが、**コミット/ロールバックの検証自体をしたいテスト**が必要になったら方式を再検討する（このADRの見直し条件）
