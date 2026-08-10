# ADR-0002: repository 層を作らない

- ステータス: Accepted
- 日付: 2026-08-08

## 背景

クリーンアーキテクチャの定番構成は handler / usecase / repository の3層で、当初案（旧R1計画）もこの形だった。
しかし本リポジトリは DB アクセスに sqlc を採用する。sqlc は `queries/*.sql` から型安全な `Queries`（メソッド群）を生成するため、**「SQLを1か所に隔離する」という repository の役割が生成物で既に満たされる**。

また、テストは go-txdb で実DBに対して行う方針のため、「モックに差し替えるためのインターフェース」も必要がない。

## 決定

repository 層は作らない。usecase が sqlc の生成する `Queries` を直接呼ぶ。
`Queries` を1対1でラップする自作インターフェースも作らない。

```
handler → usecase → generated/db（sqlc の Queries）→ DB
```

## 代替案と却下理由

| 案 | 却下理由 |
| --- | --- |
| repository インターフェースを自作し `Queries` をラップする | 生成物の上に手書きの薄皮をかぶせるだけで、**行数だけ増えて安全性は上がらない** |
| ORM ＋ repository 層 | SQL が service 側に散らばる問題への対処として repository が要る構成。sqlc なら SQL は最初から `queries/` に集まっており、前提となる問題が発生しない。SQL 資産も消える |

## 影響

- 層が1つ減り、コードの追跡が短くなる（handler → usecase → SQL）
- usecase のテストはモックではなく実DB（go-txdb）で行う——SQL・制約・トランザクションまで検証される
- **見直しの条件**: DB 以外への差し替え（外部API化・キャッシュ層など）が実際に必要になったとき、**その箇所だけ**インターフェースを切る。「いつか差し替えるかも」では切らない
