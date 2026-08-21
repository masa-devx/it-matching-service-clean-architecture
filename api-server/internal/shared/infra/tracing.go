package infra

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// SetupTracing は OpenTelemetry の送信基盤を初期化する。
//
// endpoint（OTLP の送信先・例 http://localhost:4318）が空なら**何もしない（no-op）**:
// 計装コード（span の生成・伝播）は常に動くが、記録先が無いのでコストほぼゼロで捨てられる。
// これにより本番・CI に影響を与えずにローカルだけ Jaeger へ送れる。
//
// 伝播（W3C traceparent の読み書き）は送信の有無と無関係に必要なので、無条件で設定する
func SetupTracing(ctx context.Context, endpoint string) (shutdown func(context.Context) error, err error) {
	// Next.js から届く traceparent を読み、外向きリクエストに書くための規格設定
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if endpoint == "" {
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint))
	if err != nil {
		return nil, fmt.Errorf("OTLP エクスポーターの作成に失敗: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		// Batcher: span を溜めてまとめて送る（リクエスト処理をブロックしない）
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(sdkresource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("tsunagu-api"),
		)),
	)
	otel.SetTracerProvider(provider)
	return provider.Shutdown, nil
}
