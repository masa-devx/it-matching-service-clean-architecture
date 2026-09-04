# ADR-0013: multipart は仕様に定義し、ボディのデコードだけを手書きする

- ステータス: Accepted（承認済み）
- 日付: 2026-08-31

## 背景

Phase 7（eKYC）は本人確認書類のファイルアップロードを含む。設計プラン §8 は「OpenAPI で `multipart/form-data` を定義し、oapi-codegen の生成型で受ける（要検証）」とし、「生成が弱ければアップロードのエンドポイントだけ手書きハンドラにする」判断を Phase 7 着手時に行うとしていた（§15 未決事項 7）。

#106 のスパイクで、試験的な KYC 2操作（multipart アップロード・署名付き URL 取得）を `.tsp` に定義し、パイプライン全体（TypeSpec 1.14 → OpenAPI → oapi-codegen v2 / orval v8）を実際に回して生成物を確認した:

| 層 | 結果 |
| --- | --- |
| TypeSpec → OpenAPI | `@multipartBody` + `HttpPart<File>` で正しい multipart 定義が出る（`format: binary`・required・`security: BearerAuth`） |
| 認証要否の導出（#31） | multipart 操作でも `security` が出るため、仕様からの導出はそのまま機能する |
| oapi-codegen（StrictServer） | ルーティング・ミドルウェア・**レスポンス型**は通常どおり生成。ただしリクエストは `KycDocumentsUploadRequestObject{ Body *multipart.Reader }` ＝**生ストリーム**。型付き構造体（`KycDocumentsUploadMultipartBody`）は生成されるがサーバー側の glue では使われない（デコードは handler の仕事） |
| orval（TS） | `FormData` の組み立てまで含む Fetch クライアント・`zod.instanceof(File)` つき Zod を完全生成。手書きの余地なし |

つまり「生成型で受けられるか」の答えは Yes/No ではなく**「デコード以外はすべて生成される」**だった。

## 決定

1. **KYC の操作も他と同様に `.tsp` に定義する**（スキーマ駆動を維持）。ルーティング・認証導出・レスポンス型・TS クライアント・Zod・CI の生成物差分チェックはすべて既存の仕組みに乗せる
2. **スキーマ駆動の例外は「multipart リクエストボディのデコード」の1点のみ**とする。handler が `*multipart.Reader` から `NextPart()` で part を取り出し、`http.MaxBytesReader`（サイズ上限）と `http.DetectContentType`（中身から MIME 判定）を自分で適用する
3. 生ストリームで渡されることは欠点ではなく**要件に合致**と評価する: 全体をメモリに載せる前にサイズを打ち切り、先頭バイトで MIME を判定する（§8 の手順）は、型付きデコード（全部読んでから検証）では実現できない
4. ローカルの署名付き URL は **fake-gcs-server ＋ ダミー秘密鍵での署名**で扱う。fake-gcs-server は署名・有効期限を**一切検証しない**（README 明記）ため、「ローカルで動く」ことは署名の正しさを保証しない。検証されるのは本番 GCS のみ、と認識して 7-4 のテストを設計する。URL のホストは `-public-host` で fake 側に向ける

## 代替案と却下理由

| 案 | 却下理由 |
| --- | --- |
| アップロードのエンドポイントごと仕様外の手書きハンドラ（当初の想定） | 弱いのはデコード1点だけと判明。エンドポイントごと外すと TS クライアント・Zod・認証導出・契約の差分チェックまで手放すことになり、例外が不必要に大きい |
| multipart をやめ、base64 を JSON に載せる | ボディが約1.33倍に膨らみ、ストリーム処理（読む前のサイズ打ち切り）ができない。multipart は標準の手段でありクライアント生成も揃っている |
| 署名付き URL 直アップロード（クライアント → GCS 直 PUT） | API がファイルを経由しない利点はあるが、サイズ・MIME の入口防衛が GCS 側設定頼みになり、学習題材（multipart の受け口）としても外れる。規模が要求したら再検討 |

## 影響

- 7-3（アップロード API）の handler は「`NextPart()` のループ・part の順序・サイズ上限・MIME 判定」を手書きする。**`MaxBytesReader` は strict glue が `r.MultipartReader()` を呼ぶ前に効かせる必要がある**ため、`r.Body` を包むのはミドルウェア層で行う
- 生成型 `KycDocumentsUploadMultipartBody` はサーバーでは未使用のまま残る（クライアント用）。「なぜ使わないのか」は本 ADR を根拠とする
- fake-gcs-server の署名付き URL は「動くが検証されない」ため、署名の正しさは本番疎通（7-6）で最終確認する
- 見直し条件: oapi-codegen が strict server で multipart の型付きデコードを提供したら、手書き部分の縮小を再検討する
