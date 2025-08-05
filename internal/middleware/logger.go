package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// responseWriter 自定义ResponseWriter，用于捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	// 将响应数据写入缓冲区
	w.body.Write(b)
	// 同时写入原始的ResponseWriter
	return w.ResponseWriter.Write(b)
}

// RequestLoggerMiddleware 记录所有请求参数的中间件
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		logRequest(c, bodyBytes)

		// 创建自定义的ResponseWriter来捕获响应
		responseBody := &bytes.Buffer{}
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           responseBody,
		}
		c.Writer = writer

		c.Next()

		logResponse(c, startTime, responseBody.Bytes())
	}
}

func logRequest(c *gin.Context, bodyBytes []byte) {
	requestLog := map[string]interface{}{
		"url":         c.Request.URL.String(),
		"remote_addr": c.ClientIP(),
		"headers":     c.Request.Header,
		"query":       c.Request.URL.Query(),
		"path_params": c.Params,
	}

	if len(bodyBytes) > 0 {
		bodyStr := string(bodyBytes)

		var jsonBody interface{}
		if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
			requestLog["body"] = jsonBody
		} else {
			if len(bodyStr) > 1000 {
				bodyStr = bodyStr[:1000] + "...[truncated]"
			}
			requestLog["body"] = bodyStr
		}
	}

	logrus.Infof("logRequest:%+v", requestLog)
}

func logResponse(c *gin.Context, startTime time.Time, responseBody []byte) {
	duration := time.Since(startTime)

	responseLog := map[string]interface{}{
		"status":      c.Writer.Status(),
		"duration_ms": float64(duration.Nanoseconds()) / 1e6,
	}

	if len(responseBody) > 0 {
		responseStr := string(responseBody)

		var jsonResponse interface{}
		if err := json.Unmarshal(responseBody, &jsonResponse); err == nil {
			responseLog["body"] = jsonResponse
		} else {
			if len(responseStr) > 5000 {
				responseStr = responseStr[:5000] + "...[truncated]"
			}
			responseLog["body"] = responseStr
		}
	}

	if len(c.Errors) > 0 {
		var errors []string
		for _, err := range c.Errors {
			errors = append(errors, err.Error())
		}
		responseLog["errors"] = errors
	}

	logrus.Infof("responseLog:%+v", responseLog)
}
