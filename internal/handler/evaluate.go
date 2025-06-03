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
	service     service.EvaluateService
	rawLogsRepo repository.RawLogsRepository
}

func NewEvaluateHandler(service service.EvaluateService, rawLogsRepo repository.RawLogsRepository) *EvaluateHandler {
	return &EvaluateHandler{
		service:     service,
		rawLogsRepo: rawLogsRepo,
	}
}

func (h *EvaluateHandler) BetaEvaluate(c *gin.Context) {
	var req model.BetaEvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, err.Error()))
		return
	}

	response, err := h.service.BetaEvaluate(c.Request.Context(), &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to evaluate essay")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "Internal server error"))
		return
	}

	go h.saveRawLog("/evaluate", req.JSONString(), response.JSONString())

	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}

func (h *EvaluateHandler) BetaOcrEvaluate(c *gin.Context) {
	var req model.BetaOcrEvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, err.Error()))
		return
	}

	response, err := h.service.BetaOcrEvaluate(c.Request.Context(), &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to evaluate OCR essay")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "Internal server error"))
		return
	}

	go h.saveRawLog("/evaluate/beta/ocr", req.JSONString(), response.JSONString())

	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
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
