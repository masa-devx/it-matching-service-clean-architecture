# ADR-0011: web も Cloud Run にデプロイする（Vercel 不採用・GCP 一元化）

- ステータス: Accepted（承認済み）
- 日付: 2026-08-19

## 背景

#82 で api は Cloud Run に載った。web（Next.js）のホスティング先を決める必要がある。
Next.js のデプロイ先としては Vercel が最有力（公式・Hobby 無料・Git 連携 CD）だが、
参考実装 [it-support-service](https://github.com/yamao-sys/it-support-service) は
フロントも Cloud Run（standalone + Docker + Cloud Build）で GCP に統一している。

このリポジトリの前提:

- API の接続先はサーバー側 env（`API_URL`）のみ・ブラウザ→API 直接呼び出しなし（CORS 不要判断と対）
- #81 で Go のコンテナ化（マルチステージ・distroless）は経験済み

## 決定

**web も Cloud Run にデプロイする**。Next.js は `output: 'standalone'` でビルドし、
マルチステージ Dockerfile（node:24-slim）でイメージ化して api と同じ
Artifact Registry / Cloud Run に載せる（GCP 一元化）。

あわせて Container Scanning（push 毎 $0.26）は**オフにする**
（代替: 無料の govulncheck + GitHub Dependabot。#82 の「様子見」判断をここで確定）。

## 代替案と却下理由

| 案                                        | 却下理由                                                                                                                                                                                                                                                                              |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Vercel（Hobby 無料）                      | Next.js 最適・CD ゼロ工数・完全無料で運用面では最良。しかし ①プラットフォームが GCP + Vercel に分散する ②Next.js のコンテナ化（standalone・モノレポ Docker ビルド）という学習機会を放棄する ③参考実装と構成を揃えた比較ができなくなる。**学習リポジトリとしての価値を優先**して見送り |
| Cloud Run + Buildpacks（Dockerfile なし） | 手軽だがビルドの中身がブラックボックスになり、学習目的に合わない                                                                                                                                                                                                                      |

## 影響

- 楽になること: GCP 1か所で API / web / Job / Secret / ログが完結。CD（#84）も1系統
- 引き受けるコスト:
  - web イメージは node ランタイム込みで **約200〜300MB**（api の約10倍）→ Artifact Registry 無料枠 0.5GB に対し **untagged の削除・クリーンアップポリシー（#84）が実質必須**
  - SSR のコールドスタートは API より重い（体感数秒・min-instances 0 の代償）
  - Vercel なら無料で付くプレビューデプロイ・画像最適化 CDN は無し
- **見直しの発動条件**（Vercel へ移す再検討）:
  1. 画像最適化・エッジ配信など Vercel 固有機能が必要になったとき
  2. コールドスタートの体感が実用を損なうと判断したとき
  3. Artifact Registry の容量・CD の維持コストが学習価値を上回ったとき
