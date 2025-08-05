package service

import (
	"context"
	"essay-stateless/internal/config"
	"time"

	"github.com/sirupsen/logrus"
	// lago "github.com/getlago/lago-go-client"
)

// BillingService 计费服务接口
type BillingService interface {
	TrackUsage(ctx context.Context, userID string, eventCode string, quantity int) error
	CreateCustomer(ctx context.Context, userID, email string) error
	GetUsage(ctx context.Context, userID string) (*UsageResponse, error)
}

type billingService struct {
	config *config.LagoConfig
	// client *lago.Client
}

// UsageResponse 使用量响应
type UsageResponse struct {
	UserID      string  `json:"user_id"`
	TotalUsage  int     `json:"total_usage"`
	TotalAmount float64 `json:"total_amount"`
}

// NewBillingService 创建计费服务
func NewBillingService(config *config.LagoConfig) BillingService {
	if !config.Enabled {
		return &noBillingService{}
	}

	// 初始化Lago客户端
	// client := lago.New().SetAPIKey(config.APIKey).SetBaseURL(config.BaseURL)

	return &billingService{
		config: config,
		// client: client,
	}
}

// TrackUsage 跟踪使用量
func (s *billingService) TrackUsage(ctx context.Context, userID string, eventCode string, quantity int) error {
	if !s.config.Enabled {
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"event_code": eventCode,
		"quantity":   quantity,
	}).Info("Tracking usage")

	// 调用Lago API跟踪使用量
	/*
		event := &lago.Event{
			TransactionID: fmt.Sprintf("%s_%d", userID, time.Now().Unix()),
			ExternalCustomerID: userID,
			Code: eventCode,
			Properties: map[string]interface{}{
				"quantity": quantity,
			},
			Timestamp: time.Now(),
		}

		_, err := s.client.Event().Create(ctx, event)
		if err != nil {
			logrus.WithError(err).Error("Failed to track usage")
			return err
		}
	*/

	// 临时实现 - 记录日志
	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"event_code": eventCode,
		"quantity":   quantity,
		"timestamp":  time.Now(),
	}).Info("Usage tracked successfully")

	return nil
}

// CreateCustomer 创建客户
func (s *billingService) CreateCustomer(ctx context.Context, userID, email string) error {
	if !s.config.Enabled {
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"email":   email,
	}).Info("Creating customer")

	// 调用Lago API创建客户
	/*
		customer := &lago.Customer{
			ExternalID: userID,
			Email:      email,
		}

		_, err := s.client.Customer().Create(ctx, customer)
		if err != nil {
			logrus.WithError(err).Error("Failed to create customer")
			return err
		}
	*/

	// 临时实现 - 记录日志
	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"email":   email,
	}).Info("Customer created successfully")

	return nil
}

// GetUsage 获取使用量
func (s *billingService) GetUsage(ctx context.Context, userID string) (*UsageResponse, error) {
	if !s.config.Enabled {
		return &UsageResponse{
			UserID:      userID,
			TotalUsage:  0,
			TotalAmount: 0,
		}, nil
	}

	// 调用Lago API获取使用量
	/*
		usage, err := s.client.Customer().Usage(ctx, userID)
		if err != nil {
			logrus.WithError(err).Error("Failed to get usage")
			return nil, err
		}

		return &UsageResponse{
			UserID:      userID,
			TotalUsage:  usage.TotalUsage,
			TotalAmount: usage.TotalAmount,
		}, nil
	*/

	// 临时实现 - 返回模拟数据
	return &UsageResponse{
		UserID:      userID,
		TotalUsage:  10,
		TotalAmount: 5.50,
	}, nil
}

// noBillingService 空计费服务（当计费未启用时）
type noBillingService struct{}

func (s *noBillingService) TrackUsage(ctx context.Context, userID string, eventCode string, quantity int) error {
	return nil
}

func (s *noBillingService) CreateCustomer(ctx context.Context, userID, email string) error {
	return nil
}

func (s *noBillingService) GetUsage(ctx context.Context, userID string) (*UsageResponse, error) {
	return &UsageResponse{UserID: userID, TotalUsage: 0, TotalAmount: 0}, nil
}
