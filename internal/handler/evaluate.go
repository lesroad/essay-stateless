package handler

import (
	"context"
	"net/http"
	"time"

	"essay-stateless/internal/model"
	"essay-stateless/internal/repository"
	"essay-stateless/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type EvaluateHandler struct {
	service        service.EvaluateService
	rawLogsRepo    repository.RawLogsRepository
	billingService service.BillingService
}

func NewEvaluateHandler(service service.EvaluateService, rawLogsRepo repository.RawLogsRepository, billingService service.BillingService) *EvaluateHandler {
	return &EvaluateHandler{
		service:        service,
		rawLogsRepo:    rawLogsRepo,
		billingService: billingService,
	}
}

// EvaluateStream SSE流式批改接口
func (h *EvaluateHandler) EvaluateStream(c *gin.Context) {
	var req model.EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, err.Error()))
		return
	}

	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	// 设置SSE响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	c.Writer.WriteHeader(http.StatusOK)

	// 立即刷新响应头
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	// 创建响应通道
	ch := make(chan *model.StreamEvaluateResponse, 50)

	// 启动流式评估
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.WithField("panic", r).Error("Panic in stream evaluation")
				// 发送panic错误消息（如果channel还没关闭）
				select {
				case ch <- &model.StreamEvaluateResponse{
					Type:      "error",
					Step:      "panic",
					Message:   "服务内部错误",
					Data:      &model.StreamErrorData{Error: "internal error", Step: "panic"},
					Timestamp: time.Now().Unix(),
				}:
				default:
					// channel已关闭或已满，忽略
				}
			}
		}()

		if err := h.service.EvaluateStream(c.Request.Context(), &req, ch); err != nil {
			logrus.WithError(err).Error("Failed to stream evaluate essay")
		}

		// 跟踪使用量
		if err := h.billingService.TrackUsage(c.Request.Context(), userID, "essay_evaluate_stream", 1); err != nil {
			logrus.WithError(err).Error("Failed to track usage for stream essay evaluation")
		}
	}()

	// 发送SSE数据
	for {
		select {
		case <-c.Request.Context().Done():
			logrus.WithField("user_id", userID).Info("Client disconnected from stream")
			return
		case msg, ok := <-ch:
			if !ok {
				// 通道关闭，结束
				return
			}

			data, err := msg.JSONString()
			if err != nil {
				logrus.WithError(err).Error("Failed to marshal stream response")
				continue
			}

			// 写入SSE格式数据
			if _, err := c.Writer.Write([]byte("event: message\n")); err != nil {
				logrus.WithError(err).Error("Failed to write SSE event")
				return
			}
			if _, err := c.Writer.Write([]byte("data: " + data + "\n\n")); err != nil {
				logrus.WithError(err).Error("Failed to write SSE data")
				return
			}

			// 立即刷新
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}

			// 记录日志（仅完成时）
			if msg.Type == "complete" {
				go h.saveRawLog("/evaluate/stream", req.JSONString(), data)
				return // 完成后结束
			}

			if msg.Type == "error" {
				return // 错误后结束
			}
		}
	}
}

func (h *EvaluateHandler) saveRawLog(url, request, response string) {
	log := &model.RawLogs{
		URL:        url,
		Request:    request,
		Response:   response,
		CreateTime: time.Now(),
	}

	if err := h.rawLogsRepo.Save(context.Background(), log); err != nil {
		logrus.WithError(err).Error("Failed to save raw log")
	}
}
