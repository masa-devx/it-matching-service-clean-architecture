# Tsunagu Works

企業 × IT人材（副業・フリーランスエンジニア）のビジネスマッチング。
[it-support-service](https://github.com/yamao-sys/it-support-service) を参考にした転職ポートフォリオ旗艦。
独自機能＝**信頼の設計**（エスクロー決済・連絡先マスキング・レビュー同時公開・稼働報告）。

- 詳細仕様: `仕様ドラフト.html`
- 開発ルール: `CLAUDE.md` / `.claude/rules/`
- 学習ログ: `docs/学習ログ.md`

## 構成（Stage 1: シンプル開始 → 段階的にリファクタ）

```
api/        Go（フラット構成 → R1でクリーンアーキへ）
web/        Next.js App Router（components/hooks/lib → R2/R3で features/external へ）
docs/       学習ログ
```

## ポート

| 用途 | ポート |
| --- | --- |
| PostgreSQL | 5434 |
| Go API | 8081 |
| Next.js | 3000 |

## 起動（ユーザーが実行）

```bash
# DB
docker compose up -d

# API（Phase 1 以降）
cd api && go run .

# Web（Phase 2 以降）
cd web && npm run dev
```
