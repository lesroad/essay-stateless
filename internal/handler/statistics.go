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

// StatisticsHandler 学情统计分析处理器
type StatisticsHandler struct {
	serviceV2   *appService.StatisticsServiceV2
	rawLogsRepo repository.RawLogsRepository
}

// NewStatisticsHandler 创建学情统计分析处理器
func NewStatisticsHandler(serviceV2 *appService.StatisticsServiceV2, rawLogsRepo repository.RawLogsRepository) *StatisticsHandler {
	return &StatisticsHandler{
		serviceV2:   serviceV2,
		rawLogsRepo: rawLogsRepo,
	}
}

// AnalyzeClassStatistics 班级学情统计分析接口
func (h *StatisticsHandler) AnalyzeClassStatistics(c *gin.Context) {
	var req model.ClassStatisticsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Failed to bind class statistics request")
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "请求参数格式错误: "+err.Error()))
		return
	}

	// 参数验证
	if len(req) == 0 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "学生数据不能为空"))
		return
	}

	if len(req) > 100 {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "学生数量不能超过100人"))
		return
	}

	// 调用服务层进行分析
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	response, err := h.serviceV2.AnalyzeClassStatistics(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "统计分析失败: "+err.Error()))
		return
	}

	// 异步保存日志
	go func() {
		h.saveRawLog("/statistics/class", req.JSONString(), response.JSONString())
	}()

	c.JSON(http.StatusOK, model.NewSuccessResponse(response))
}

// saveRawLog 保存原始日志
func (h *StatisticsHandler) saveRawLog(url, request, response string) {
	log := &model.RawLogs{
		URL:        url,
		Request:    request,
		Response:   response,
		CreateTime: time.Now(),
	}

	if err := h.rawLogsRepo.Save(context.Background(), log); err != nil {
		logrus.WithError(err).Error("Failed to save raw log for statistics")
	}
}
