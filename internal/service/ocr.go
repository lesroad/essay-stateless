package service

import (
	"context"
	"encoding/json"
	"essay-stateless/internal/config"
	"essay-stateless/internal/model"
	"essay-stateless/pkg/httpclient"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type OcrService interface {
	DefaultOcr(ctx context.Context, provider, imageType string, req *model.DefaultOcrRequest) (*model.DefaultOcrResponse, error)
	TitleOcr(ctx context.Context, provider, imageType string, req *model.TitleOcrRequest) (*model.TitleOcrResponse, error)
}

type ocrService struct {
	config       *config.OCRConfig
	httpClient   *httpclient.Client
	openaiClient *openai.Client
}

func NewOcrService(config *config.OCRConfig) OcrService {
	// 创建 OpenAI 客户端配置用于 ARK
	arkConfig := openai.DefaultConfig(config.ArkAPIKey)
	arkConfig.BaseURL = config.ArkBaseURL
	arkClient := openai.NewClientWithConfig(arkConfig)

	return &ocrService{
		config:       config,
		httpClient:   httpclient.New(),
		openaiClient: arkClient,
	}
}

func (s *ocrService) DefaultOcr(ctx context.Context, provider, imageType string, req *model.DefaultOcrRequest) (*model.DefaultOcrResponse, error) {
	if provider == "" {
		provider = s.config.DefaultProvider
	}

	switch provider {
	case "bee":
		return s.beeDefaultOcr(ctx, imageType, req)
	case "ark":
		return s.arkDefaultOcr(ctx, req)
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
	case "ark":
		return s.arkTitleOcr(ctx, req)
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
		return nil, fmt.Errorf("OCR服务返回错误!: %s, %d", ocrResp.Msg, ocrResp.Code)
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

// ARK OCR响应结构体
type ArkOcrResponse struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// ARK Default OCR实现
func (s *ocrService) arkDefaultOcr(ctx context.Context, req *model.DefaultOcrRequest) (*model.DefaultOcrResponse, error) {
	if len(req.Images) == 0 {
		return &model.DefaultOcrResponse{
			Title:   nil,
			Content: "",
		}, nil
	}

	var allContent []string
	for _, imageURL := range req.Images {
		content, err := s.callArkOCR(ctx, imageURL, false)
		if err != nil {
			return nil, err
		}
		if content != "" {
			allContent = append(allContent, content)
		}
	}

	return &model.DefaultOcrResponse{
		Title:   nil,
		Content: strings.Join(allContent, "\n"),
	}, nil
}

// ARK Title OCR实现
func (s *ocrService) arkTitleOcr(ctx context.Context, req *model.TitleOcrRequest) (*model.TitleOcrResponse, error) {
	if len(req.Images) == 0 {
		return &model.TitleOcrResponse{
			Title:   "",
			Content: "",
		}, nil
	}

	// 处理第一张图片作为标题
	var title, content string
	for i, imageURL := range req.Images {
		if i == 0 {
			// 第一张图片，解析标题和内容
			result, err := s.callArkOCR(ctx, imageURL, true)
			if err != nil {
				return nil, err
			}
			var arkResp ArkOcrResponse
			if err := json.Unmarshal([]byte(result), &arkResp); err != nil {
				// 如果解析失败，将整个结果作为内容
				logrus.Errorf("failed to parse ark, err:%v", err)
				title = ""
				content = result
			} else {
				title = arkResp.Title
				content = arkResp.Text
			}
		} else {
			// 后续图片只作为内容
			result, err := s.callArkOCR(ctx, imageURL, false)
			if err != nil {
				return nil, err
			}
			if content != "" {
				content += "\n"
			}
			content += result
		}
	}

	return &model.TitleOcrResponse{
		Title:   title,
		Content: content,
	}, nil
}

// 调用ARK OCR API
func (s *ocrService) callArkOCR(ctx context.Context, imageURL string, withTitle bool) (string, error) {
	// 构建提示词
	prompt := `请阅读这张学生的作文图片，并提取以下信息:
1. 作文标题，一般来说是第一面的第一行，也有可能是作文题目的要求(如丁组xxx)
2. 作文正文内容(不需要题号或开头提示语),精确识别每个段落，注意图片种段落开头会缩进,识别后以\n作为段落换行符
3. 不需要输出别的内容，输出为json对象
4. 如果不能识别，也要返回信息，将作文标题、作文正文置为空字符串
请将结果以 JSON 格式输出，如下格式:
{"title": "(作文标题)","text": "(作文正文)"}`

	if !withTitle {
		prompt = `请阅读这张学生的作文图片，并提取作文正文内容(不需要题号或开头提示语),精确识别每个段落，注意图片中段落开头会缩进,识别后以\n作为段落换行符。
如果不能识别，返回空字符串。只返回文本内容，不要JSON格式。`
	}

	// 构建消息
	messages := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleUser,
			MultiContent: []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL: imageURL,
					},
				},
				{
					Type: openai.ChatMessagePartTypeText,
					Text: prompt,
				},
			},
		},
	}

	// 创建请求
	request := openai.ChatCompletionRequest{
		Model:    s.config.ArkModel,
		Messages: messages,
	}

	// 调用API
	response, err := s.openaiClient.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", fmt.Errorf("ARK OCR API调用失败: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("ARK OCR API返回空响应")
	}

	return response.Choices[0].Message.Content, nil
}
