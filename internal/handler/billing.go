package handler

import (
	"net/http"

	"essay-stateless/internal/model"
	"essay-stateless/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type BillingHandler struct {
	billingService service.BillingService
}

func NewBillingHandler(billingService service.BillingService) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
	}
}

// CreateCustomer 创建客户
func (h *BillingHandler) CreateCustomer(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Email  string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, err.Error()))
		return
	}

	err := h.billingService.CreateCustomer(c.Request.Context(), req.UserID, req.Email)
	if err != nil {
		logrus.WithError(err).Error("Failed to create customer")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "Internal server error"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(map[string]string{
		"message": "Customer created successfully",
	}))
}

// GetUsage 获取用户使用量
func (h *BillingHandler) GetUsage(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, "user_id is required"))
		return
	}

	usage, err := h.billingService.GetUsage(c.Request.Context(), userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to get usage")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "Internal server error"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(usage))
}

// TrackUsage 手动跟踪使用量（用于测试）
func (h *BillingHandler) TrackUsage(c *gin.Context) {
	var req struct {
		UserID    string `json:"user_id" binding:"required"`
		EventCode string `json:"event_code" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, err.Error()))
		return
	}

	err := h.billingService.TrackUsage(c.Request.Context(), req.UserID, req.EventCode, req.Quantity)
	if err != nil {
		logrus.WithError(err).Error("Failed to track usage")
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, "Internal server error"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(map[string]string{
		"message": "Usage tracked successfully",
	}))
}
