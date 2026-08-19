# 本番用イメージ（マルチステージ）。ビルドコンテキストはリポジトリルート:
#   docker build -f api-server/Dockerfile -t tsunagu-api .
# 1イメージに server / migrate の2バイナリを同居させ、
# Cloud Run サービスは既定の /server・Job は command 上書きで /migrate を使う
# ＝「テストを通った同じイメージ」が両用途に行く（イメージ＝不変の成果物）

FROM golang:1.26 AS builder
WORKDIR /src

# 依存だけ先に取得（go.mod / go.sum が変わらない限りこのレイヤーはキャッシュされる）
COPY api-server/go.mod api-server/go.sum ./
RUN go mod download

COPY api-server/ ./
# CGO_ENABLED=0: C ライブラリに依存しない静的バイナリにする（distroless/static で動く条件）。
# -trimpath / -ldflags "-s -w": ビルドパスとデバッグ情報を落としてサイズ削減
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# distroless/static: シェルもパッケージマネージャも無い最小イメージ（攻撃面とサイズの最小化）。
# nonroot タグで非 root 実行
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/server /server
COPY --from=builder /out/migrate /migrate
# migrations は api-server モジュールの外にあるため go:embed できない → ファイルとして同梱
# （cmd/migrate の MIGRATIONS_DIR 既定値 /migrations/ddl と対応）
COPY migrations/ddl /migrations/ddl

EXPOSE 8082
ENTRYPOINT ["/server"]
