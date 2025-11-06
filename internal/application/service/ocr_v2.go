package service

import (
	"context"
	"essay-stateless/internal/config"
	"essay-stateless/internal/domain/evaluate"
	"essay-stateless/internal/domain/ocr"
	"essay-stateless/internal/model"
	"strings"
)

// OcrServiceV2 新版OCR服务（基于DDD架构）
type OcrServiceV2 struct {
	config         *config.OCRConfig
	beeProvider    *ocr.BeeProvider
	arkProvider    *ocr.ArkProvider
	contentCleaner *evaluate.ContentCleaner
}

// NewOcrServiceV2 创建新版OCR服务
func NewOcrServiceV2(config *config.OCRConfig) *OcrServiceV2 {
	return &OcrServiceV2{
		config:         config,
		beeProvider:    ocr.NewBeeProvider(config.BeeAPI, config.XAppKey, config.XAppSecret),
		arkProvider:    ocr.NewArkProvider(config.ArkAPIKey, config.ArkBaseURL, config.ArkModel),
		contentCleaner: evaluate.NewContentCleaner(),
	}
}

// TitleOcr 带标题OCR识别
func (s *OcrServiceV2) TitleOcr(ctx context.Context, provider, imageType string, req *model.TitleOcrRequest) (*model.TitleOcrResponse, error) {
	if provider == "" {
		provider = s.config.DefaultProvider
	}

	switch provider {
	case "bee":
		return s.beeTitleOcr(ctx, imageType, req)
	case "ark":
		return s.arkTitleOcr(ctx, req)
	default:
		return s.beeTitleOcr(ctx, imageType, req)
	}
}

// beeTitleOcr 使用Bee提供商进行OCR识别
func (s *OcrServiceV2) beeTitleOcr(ctx context.Context, imageType string, req *model.TitleOcrRequest) (*model.TitleOcrResponse, error) {
	leftType := "all"
	if req.LeftType != nil {
		leftType = *req.LeftType
	}

	results, err := s.beeProvider.RecognizeImages(ctx, req.Images, imageType, leftType)
	if err != nil {
		return nil, err
	}

	var title, content string
	if len(results) > 0 {
		title = results[0]
		if len(results) > 1 {
			content = strings.Join(results[1:], "\n")
		}
	}

	return &model.TitleOcrResponse{
		Title:   title,
		Content: s.contentCleaner.Clean(content),
	}, nil
}

// arkTitleOcr 使用ARK提供商进行OCR识别
func (s *OcrServiceV2) arkTitleOcr(ctx context.Context, req *model.TitleOcrRequest) (*model.TitleOcrResponse, error) {
	title, content, err := s.arkProvider.RecognizeWithTitle(ctx, req.Images)
	if err != nil {
		return nil, err
	}

	return &model.TitleOcrResponse{
		Title:   title,
		Content: s.contentCleaner.Clean(content),
	}, nil
}

