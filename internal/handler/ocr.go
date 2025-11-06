package handler

import (
	"context"
	"net/http"
	"time"

	appService "essay-stateless/internal/application/service"
	"essay-stateless/internal/model"
	"essay-stateless/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OcrHandler struct {
	serviceV2   *appService.OcrServiceV2
	rawLogsRepo repository.RawLogsRepository
}

func NewOcrHandler(serviceV2 *appService.OcrServiceV2, rawLogsRepo repository.RawLogsRepository) *OcrHandler {
	return &OcrHandler{
		serviceV2:   serviceV2,
		rawLogsRepo: rawLogsRepo,
	}
}

// TitleOcr 带标题OCR识别
// @param provider OCR的提供者, textin or bee
// @param imgType  OCR识别类型, url or base64
// @param req      OCR识别请求
// @return OCR 识别结果，包含标题
func (h *OcrHandler) TitleOcr(c *gin.Context) {
	provider := c.Param("provider")
	imgType := c.Param("imgType")

	var req model.TitleOcrRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, err.Error()))
		return
	}

	response, err := h.serviceV2.TitleOcr(c.Request.Context(), provider, imgType, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to perform title OCR")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "Internal server error"))
		return
	}

	// 记录日志
	go h.saveRawLog("/sts/ocr/title/"+provider+"/"+imgType, req.JSONString(), response)

	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}

func (h *OcrHandler) saveRawLog(url, request string, response any) {
	responseStr := ""
	if resp, ok := response.(*model.DefaultOcrResponse); ok {
		if data, err := resp.JSONString(); err == nil {
			responseStr = data
		}
	} else if resp, ok := response.(*model.TitleOcrResponse); ok {
		if data, err := resp.JSONString(); err == nil {
			responseStr = data
		}
	}

	log := &model.RawLogs{
		URL:        url,
		Request:    request,
		Response:   responseStr,
		CreateTime: time.Now(),
	}

	if err := h.rawLogsRepo.Save(context.Background(), log); err != nil {
		logrus.WithError(err).Error("Failed to save raw log")
	}
}
