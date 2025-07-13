package trace

import (
	"context"
	"essay-stateless/internal/config"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

func Init(config config.TraceConfig) (func(), error) {
	// 如果没有配置 endpoint，就不设置 exporter（使用默认的noop）
	if config.Endpoint == "" {
		return func() {}, nil
	}

	// 创建 Resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建 OTLP HTTP exporter（更通用，支持Jaeger、Grafana Tempo等）
	exp, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(config.Endpoint),
		otlptracehttp.WithInsecure(), // 开发环境使用，生产环境建议移除
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create otlp http exporter: %w", err)
	}

	// 创建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(tp)

	// 设置全局 Propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		b3.New(),
		propagation.Baggage{},
		propagation.TraceContext{},
	))

	// 返回清理函数
	return func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			fmt.Printf("Error shutting down tracer provider: %v\n", err)
		}
	}, nil
}

func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		traceID := span.SpanContext().TraceID().String()
		c.Header("X-Trace-Id", traceID)
		c.Next()
	}
}
