package ocr

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

// ArkProvider ARK OCR提供商
type ArkProvider struct {
	client *openai.Client
	model  string
}

// NewArkProvider 创建ARK OCR提供商
func NewArkProvider(apiKey, baseURL, model string) *ArkProvider {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	client := openai.NewClientWithConfig(config)

	return &ArkProvider{
		client: client,
		model:  model,
	}
}

// RecognizeWithTitle 识别图片并提取标题和内容
func (p *ArkProvider) RecognizeWithTitle(ctx context.Context, images []string) (string, string, error) {
	if len(images) == 0 {
		return "", "", nil
	}

	var title, content string

	for i, imageURL := range images {
		if i == 0 {
			// 第一张图片，解析标题和内容
			result, err := p.callAPI(ctx, imageURL, true)
			if err != nil {
				return "", "", err
			}
			logrus.Infof("ARK OCR识别结果: %s", result)
			var arkResp ArkOcrResponse
			if err := json.Unmarshal([]byte(result), &arkResp); err != nil {
				// 如果解析失败，将整个结果作为内容
				logrus.Errorf("failed to parse ARK response, err:%v", err)
				title = ""
				content = result
			} else {
				title = arkResp.Title
				content = arkResp.Text
			}
		} else {
			// 后续图片只作为内容
			result, err := p.callAPI(ctx, imageURL, false)
			if err != nil {
				return "", "", err
			}
			if content != "" {
				content += "\n"
			}
			content += result
		}
	}

	return title, content, nil
}

// callAPI 调用ARK OCR API
func (p *ArkProvider) callAPI(ctx context.Context, imageURL string, withTitle bool) (string, error) {
	prompt := p.buildPrompt(withTitle)

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

	request := openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: messages,
	}

	response, err := p.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", fmt.Errorf("ARK OCR API调用失败: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("ARK OCR API返回空响应")
	}

	return response.Choices[0].Message.Content, nil
}

// buildPrompt 构建提示词
func (p *ArkProvider) buildPrompt(withTitle bool) string {
	if withTitle {
		return `请阅读这张学生的作文图片，并提取以下信息:
1. 作文标题，一般来说是第一面的第一行，也有可能是作文题目的要求(如丁组xxx)
2. 作文正文内容(不需要题号或开头提示语),精确识别每个段落，注意图片中段落开头会缩进

输出要求：
- 必须返回标准的单行JSON格式，不要换行
- text字段中的段落换行请用 \n 表示（两个字符：反斜杠和n）
- 不要在JSON中包含真实的换行符
- 如果不能识别，将作文标题、作文正文置为空字符串

示例输出：
{"title":"我的梦想","text":"第一段内容。\n第二段内容。\n第三段内容。"}

请严格按照上述格式输出JSON，确保是有效的单行JSON字符串。`
	}

	return `请阅读这张学生的作文图片，并提取作文正文内容(不需要题号或开头提示语),精确识别每个段落，注意图片中段落开头会缩进,识别后以\n作为段落换行符。
如果不能识别，返回空字符串。只返回文本内容，不要JSON格式。`
}

// ArkOcrResponse ARK OCR响应
type ArkOcrResponse struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}
