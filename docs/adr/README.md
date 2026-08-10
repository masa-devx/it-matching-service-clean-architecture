# ADR（Architecture Decision Record）

アーキテクチャ上の**決定**を「1判断 = 1ファイル」で記録する。
全体像や設計の解説は [後継リポジトリ設計プラン](../後継リポジトリ設計プラン.md) が担い、ここには**個々の判断の経緯と却下した代替案**を残す。

## いつ書くか

1. **保留事項を決めたとき**（設計プラン §13 のリストが対象）
2. **構造に影響する技術・方針を選んだとき**（層の切り方・生成方式・デプロイ構成など）
3. **過去の決定を覆すとき**（下記の不変ルールに従う）

画面の文言やライブラリのマイナーバージョンなど、後から低コストで変えられるものは書かない。
**「戻すのが高くつく判断」だけを書く。**

## 運用ルール

- [template.md](template.md) の構成（ステータス / 背景 / 決定 / 代替案と却下理由 / 影響）に従う
- ファイル名は `NNNN-英語ケバブケース.md`（連番は欠番を作らない）
- **一度 Accepted にした ADR は書き換えない（不変）**。決定を変えるときは新しい ADR を起こし、
  旧 ADR のステータスだけを `Superseded by ADR-YYYY` に更新する
  — こうすることで「いつ・何を根拠に判断が変わったか」の履歴が消えない
- ステータスの語彙: `Proposed`（提案中） / `Accepted`（採用） / `Superseded`（新しい決定に置き換え） / `Deprecated`（廃止）

## 一覧

| # | タイトル | ステータス |
| --- | --- | --- |
| [0001](0001-keep-postgresql.md) | PostgreSQL を維持する | Accepted |
| [0002](0002-no-repository-layer.md) | repository 層を作らない | Accepted |
| [0003](0003-viewpoint-first-directories.md) | ディレクトリは「視点→層」で分割する | Accepted |
| [0004](0004-no-runtime-validation.md) | フロントエンドでランタイム検証をしない | Accepted |
| [0005](0005-separate-ports-containers.md) | ポート・コンテナ名を tsunagu-works と分離する | Accepted |
| [0006](0006-single-frontend-after-only.md) | このリポジトリを After 専用にする（フロント1アプリ・スナップショット撤去） | Accepted |
| [0007](0007-form-zod-from-generated.md) | フォーム検証は生成Zodの加工で行う（案1＋案4） | Accepted |
