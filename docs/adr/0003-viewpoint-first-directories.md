# ADR-0003: ディレクトリは「視点→層」で分割する

- ステータス: Accepted
- 日付: 2026-08-08

## 背景

本サービスには company（企業）と talent（人材）の2つの視点があり、同じリソース（案件・応募）でも見える世界と許される操作が全く違う。
バックエンドのディレクトリをどの軸で割るかに、「層が先」と「視点が先」の2案があった:

```
❌ 層が先                    ✅ 視点が先（採用）
handler/                     company/
├─ company_project.go        ├─ handler/
└─ talent_project.go         ├─ usecase/
usecase/                     └─ validator/
├─ company_project.go        talent/
└─ talent_project.go         ├─ handler/
                             └─ usecase/
```

## 決定

`internal/{company, talent, shared}` の**視点が先**の分割を採る。各視点の下に handler / usecase / validator の層を置き、両視点で共有するもの（状態遷移表・認証・インフラ）だけを `shared/` に置く。

## 代替案と却下理由

| 案 | 却下理由 |
| --- | --- |
| 層が先（handler/ usecase/ の下にファイルで視点分け） | 将来デプロイ単位を分けたくなったときディレクトリ大移動になる。認可の境界とディレクトリ境界が一致しない |
| フェーズで割る（登録 / 業務） | 本サービスは**企業と人材で見える世界がそもそも違う**ため、視点で割るほうが境界が安定する |

## 影響

- 将来 `cmd/company-server` / `cmd/talent-server` に分けたくなっても、ディレクトリを移動せずに済む
- フロント（`apps/company` / `apps/talent`）と境界の切り方が揃う
- ロール認可がパスプレフィックス（`/company/*` / `/talent/*`）×ミドルウェアで一律に決まり、**ハンドラごとのロール判定の書き忘れが構造的に起きない**
- `internal/company` ⇔ `internal/talent` の相互 import は depguard で禁止する
- **見直しの条件**: 契約・メッセージ等「同じ処理を実行者ロールだけ変えて呼ぶ」機能をスコープに加えるとき、`shared/usecase/` の追加を検討する（現スコープでは不要）
