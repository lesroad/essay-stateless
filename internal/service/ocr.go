package service

import (
	"context"
	"essay-stateless/internal/config"
	"essay-stateless/internal/model"
	"essay-stateless/pkg/httpclient"
	"fmt"
	"strings"
)

type OcrService interface {
	DefaultOcr(ctx context.Context, provider, imageType string, req *model.DefaultOcrRequest) (*model.DefaultOcrResponse, error)
	TitleOcr(ctx context.Context, provider, imageType string, req *model.TitleOcrRequest) (*model.TitleOcrResponse, error)
}

type ocrService struct {
	config     *config.OCRConfig
	httpClient *httpclient.Client
}

func NewOcrService(config *config.OCRConfig) OcrService {
	return &ocrService{
		config:     config,
		httpClient: httpclient.New(),
	}
}

func (s *ocrService) DefaultOcr(ctx context.Context, provider, imageType string, req *model.DefaultOcrRequest) (*model.DefaultOcrResponse, error) {
	if provider == "" {
		provider = s.config.DefaultProvider
	}

	switch provider {
	case "bee":
		return s.beeDefaultOcr(ctx, imageType, req)
	default:
		return s.beeDefaultOcr(ctx, imageType, req)
	}
}

func (s *ocrService) TitleOcr(ctx context.Context, provider, imageType string, req *model.TitleOcrRequest) (*model.TitleOcrResponse, error) {
	if provider == "" {
		provider = s.config.DefaultProvider
	}

	switch provider {
	case "bee":
		return s.beeTitleOcr(ctx, imageType, req)
	default:
		return s.beeTitleOcr(ctx, imageType, req)
	}
}

func (s *ocrService) beeDefaultOcr(ctx context.Context, imageType string, req *model.DefaultOcrRequest) (*model.DefaultOcrResponse, error) {
	leftType := "all"
	if req.LeftType != nil {
		leftType = *req.LeftType
	}

	results, err := s.ocrImages(ctx, req.Images, imageType, leftType)
	if err != nil {
		return nil, err
	}

	content := strings.Join(results, "\n")
	return &model.DefaultOcrResponse{
		Title:   nil, 
		Content: content,
	}, nil
}

func (s *ocrService) beeTitleOcr(ctx context.Context, imageType string, req *model.TitleOcrRequest) (*model.TitleOcrResponse, error) {
	leftType := "all"
	if req.LeftType != nil {
		leftType = *req.LeftType
	}

	results, err := s.ocrImages(ctx, req.Images, imageType, leftType)
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
		Content: content,
	}, nil
}

func (s *ocrService) ocrImages(ctx context.Context, images []string, imageType, leftType string) ([]string, error) {
	headers := map[string]string{
		"x-app-key":    s.config.XAppKey,
		"x-app-secret": s.config.XAppSecret,
		"Content-Type": "application/json",
	}

	var results []string

	imageParam := "image_" + imageType

	for _, image := range images {
		result, err := s.ocrOne(ctx, image, imageParam, leftType, headers)
		if err != nil {
			return nil, err
		}
		results = append(results, result...)
	}

	return results, nil
}

func (s *ocrService) ocrOne(ctx context.Context, image, imageParam, leftType string, headers map[string]string) ([]string, error) {
	body := map[string]interface{}{
		imageParam: image,
	}

	var ocrResp BeeOcrResponse
	if err := s.httpClient.PostWithHeaders(ctx, s.config.BeeAPI, body, &ocrResp, headers); err != nil {
		return nil, fmt.Errorf("OCR服务调用失败: %w", err)
	}

	if ocrResp.Code != 0 {
		return nil, fmt.Errorf("OCR服务返回错误: %s", ocrResp.Msg)
	}

	return s.processOcrResponse(ocrResp.Data, leftType)
}

func (s *ocrService) processOcrResponse(data *BeeOcrData, leftType string) ([]string, error) {
	if data == nil {
		return []string{}, nil
	}

	// 获取需要排除的类型
	var exclude int = -1
	switch leftType {
	case "handwriting":
		exclude = 0 
	case "print":
		exclude = 1 
	default:
		exclude = -1 
	}

	excludes := make(map[int]bool)
	if exclude != -1 {
		for _, line := range data.Lines {
			if line.Handwritten == exclude {
				excludes[line.AreaIndex] = true
			}
		}
	}

	var results []string
	for _, area := range data.Areas {
		if !excludes[area.Index] && strings.TrimSpace(area.Text) != "" {
			results = append(results, area.Text)
		}
	}

	return results, nil
}

type BeeOcrResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data *BeeOcrData `json:"data"`
}

type BeeOcrData struct {
	Lines []BeeOcrLine `json:"lines"`
	Areas []BeeOcrArea `json:"areas"`
}

type BeeOcrLine struct {
	Handwritten int `json:"handwritten"`
	AreaIndex   int `json:"area_index"`
}

type BeeOcrArea struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}
