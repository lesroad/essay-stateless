package ocr

import (
	"context"
	"essay-stateless/pkg/httpclient"
	"fmt"
	"strings"
)

// BeeProvider Bee OCR提供商
type BeeProvider struct {
	apiURL    string
	appKey    string
	appSecret string
	client    *httpclient.Client
}

// NewBeeProvider 创建Bee OCR提供商
func NewBeeProvider(apiURL, appKey, appSecret string) *BeeProvider {
	return &BeeProvider{
		apiURL:    apiURL,
		appKey:    appKey,
		appSecret: appSecret,
		client:    httpclient.New(),
	}
}

// RecognizeImages 识别多张图片
func (p *BeeProvider) RecognizeImages(ctx context.Context, images []string, imageType, leftType string) ([]string, error) {
	headers := map[string]string{
		"x-app-key":    p.appKey,
		"x-app-secret": p.appSecret,
		"Content-Type": "application/json",
	}

	var results []string
	imageParam := "image_" + imageType

	for _, image := range images {
		result, err := p.recognizeOne(ctx, image, imageParam, leftType, headers)
		if err != nil {
			return nil, err
		}
		results = append(results, result...)
	}

	return results, nil
}

// recognizeOne 识别单张图片
func (p *BeeProvider) recognizeOne(ctx context.Context, image, imageParam, leftType string, headers map[string]string) ([]string, error) {
	body := map[string]any{
		imageParam: image,
	}

	var ocrResp BeeOcrResponse
	if err := p.client.PostWithHeaders(ctx, p.apiURL, body, &ocrResp, headers); err != nil {
		return nil, fmt.Errorf("Bee OCR服务调用失败: %w", err)
	}

	if ocrResp.Code != 0 {
		return nil, fmt.Errorf("Bee OCR服务返回错误: %s, 错误码: %d", ocrResp.Msg, ocrResp.Code)
	}

	return p.processResponse(ocrResp.Data, leftType)
}

// processResponse 处理OCR响应
func (p *BeeProvider) processResponse(data *BeeOcrData, leftType string) ([]string, error) {
	if data == nil {
		return []string{}, nil
	}

	// 获取需要排除的类型
	exclude := p.getExcludeType(leftType)

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

// getExcludeType 获取需要排除的类型
func (p *BeeProvider) getExcludeType(leftType string) int {
	switch leftType {
	case "handwriting":
		return 0 // 排除打印体
	case "print":
		return 1 // 排除手写体
	default:
		return -1 // 不排除
	}
}

// BeeOcrResponse Bee OCR响应
type BeeOcrResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data *BeeOcrData `json:"data"`
}

// BeeOcrData Bee OCR数据
type BeeOcrData struct {
	Lines []BeeOcrLine `json:"lines"`
	Areas []BeeOcrArea `json:"areas"`
}

// BeeOcrLine Bee OCR行
type BeeOcrLine struct {
	Handwritten int `json:"handwritten"`
	AreaIndex   int `json:"area_index"`
}

// BeeOcrArea Bee OCR区域
type BeeOcrArea struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}
