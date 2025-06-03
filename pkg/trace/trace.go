package trace

import (
	"essay-stateless/internal/config"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

func Init(config config.TraceConfig) (func(), error) {
	// 初始化OpenTelemetry
	// 这里需要根据具体的追踪服务进行配置
	return func() {}, nil
}

func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		traceID := span.SpanContext().TraceID().String()
		c.Header("X-Trace-Id", traceID)
		c.Next()
	}
}
