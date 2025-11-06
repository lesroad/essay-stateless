package service

import (
	"context"
	"essay-stateless/internal/config"
	"essay-stateless/internal/domain/evaluate"
	"essay-stateless/internal/model"

	"github.com/sirupsen/logrus"
)

// EvaluateServiceV2 新版评估服务（基于DDD架构）
type EvaluateServiceV2 struct {
	config *config.EvaluateConfig

	// 领域对象
	contentCleaner    *evaluate.ContentCleaner
	clientsFactory    *evaluate.APIClientsFactory
	streamCoordinator *evaluate.StreamCoordinator
	responseProcessor *evaluate.ResponseProcessor
}

// NewEvaluateServiceV2 创建新版评估服务
func NewEvaluateServiceV2(config *config.EvaluateConfig) *EvaluateServiceV2 {
	return &EvaluateServiceV2{
		config:            config,
		contentCleaner:    evaluate.NewContentCleaner(),
		clientsFactory:    evaluate.NewAPIClientsFactory(&config.API),
		streamCoordinator: evaluate.NewStreamCoordinator(),
		responseProcessor: evaluate.NewResponseProcessor(),
	}
}

// EvaluateStream 流式批改评估
//
// DDD架构实现，包含：
// - 内容清理（ContentCleaner）
// - API客户端工厂（APIClientsFactory）
// - 流式协调器（StreamCoordinator）
// - 响应处理器（ResponseProcessor）
// - 重试执行器（RetryExecutor）
// - 分数计算器（ScoreCalculator）
// - 位置计算器（PositionCalculator）
func (s *EvaluateServiceV2) EvaluateStream(ctx context.Context, req *model.EvaluateRequest, ch chan<- *model.StreamEvaluateResponse) error {
	logrus.Info("EvaluateServiceV2: 开始流式作文批改")

	// 1. 清理内容（使用领域对象）
	req.Content = s.contentCleaner.Clean(req.Content)
	logrus.Infof("清理后作文：%s", req.Content)

	// 2. 准备模型版本信息
	modelVersion := model.ModelVersion{
		Name:    s.config.ModelVersion.Name,
		Version: s.config.ModelVersion.Version,
	}

	// 3. 使用流式协调器进行评估（ResponseProcessor在内部被调用）
	err := s.streamCoordinator.CoordinateEvaluation(ctx, req, ch, s.clientsFactory, modelVersion)
	if err != nil {
		logrus.Errorf("评估协调失败: %v", err)
		return err
	}

	logrus.Info("EvaluateServiceV2: 流式作文批改完成")
	return nil
}
