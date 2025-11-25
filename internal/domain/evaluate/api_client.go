package evaluate

import (
	"context"
	dto_evaluate "essay-stateless/internal/dto/evaluate"
	"essay-stateless/internal/model"
	"essay-stateless/pkg/httpclient"
	"fmt"

	"github.com/jinzhu/copier"
)

// APIClient 评估API客户端接口
type APIClient interface {
	Call(ctx context.Context, url string, request *model.EvaluateRequest) (any, error)
}

// BaseAPIClient 基础API客户端
type BaseAPIClient struct {
	client *httpclient.Client
	apiURL string
}

// NewBaseAPIClient 创建基础API客户端
func NewBaseAPIClient(apiURL string) *BaseAPIClient {
	return &BaseAPIClient{
		client: httpclient.New(),
		apiURL: apiURL,
	}
}

// EssayInfoClient 作文基本信息客户端
type EssayInfoClient struct {
	*BaseAPIClient
}

func NewEssayInfoClient(apiURL string) *EssayInfoClient {
	return &EssayInfoClient{
		BaseAPIClient: NewBaseAPIClient(apiURL),
	}
}

func (c *EssayInfoClient) GetEssayInfo(ctx context.Context, req *model.EvaluateRequest) (*dto_evaluate.APIEssayInfo, error) {
	data := map[string]any{
		"title": req.Title,
		"essay": req.Content,
	}

	var response dto_evaluate.APIEssayInfo
	if err := c.client.Post(ctx, c.apiURL, data, &response); err != nil {
		return nil, fmt.Errorf("获取作文基本信息失败: %w", err)
	}
	return &response, nil
}

// WordSentenceClient 词句评估客户端
type WordSentenceClient struct {
	*BaseAPIClient
}

func NewWordSentenceClient(apiURL string) *WordSentenceClient {
	return &WordSentenceClient{
		BaseAPIClient: NewBaseAPIClient(apiURL),
	}
}

func (c *WordSentenceClient) Evaluate(ctx context.Context, essay map[string]any) (*dto_evaluate.APIWordSentence, error) {
	var response dto_evaluate.APIWordSentence
	if err := c.client.Post(ctx, c.apiURL, essay, &response); err != nil {
		return nil, fmt.Errorf("词句评估失败: %w", err)
	}
	return &response, nil
}

// GrammarClient 语法检查客户端
type GrammarClient struct {
	*BaseAPIClient
}

func NewGrammarClient(apiURL string) *GrammarClient {
	return &GrammarClient{
		BaseAPIClient: NewBaseAPIClient(apiURL),
	}
}

func (c *GrammarClient) Check(ctx context.Context, essay map[string]any) (*dto_evaluate.APIGrammarInfo, error) {
	var response dto_evaluate.APIGrammarInfo
	if err := c.client.Post(ctx, c.apiURL, essay, &response); err != nil {
		return nil, fmt.Errorf("语法检查失败: %w", err)
	}
	return &response, nil
}

// FluencyClient 流畅度评估客户端
type FluencyClient struct {
	*BaseAPIClient
}

func NewFluencyClient(apiURL string) *FluencyClient {
	return &FluencyClient{
		BaseAPIClient: NewBaseAPIClient(apiURL),
	}
}

func (c *FluencyClient) Evaluate(ctx context.Context, essay map[string]any) (*dto_evaluate.APIFluency, error) {
	var response dto_evaluate.APIFluency
	if err := c.client.Post(ctx, c.apiURL, essay, &response); err != nil {
		return nil, fmt.Errorf("流畅度评估失败: %w", err)
	}
	return &response, nil
}

// OverallClient 总体评价客户端
type OverallClient struct {
	*BaseAPIClient
}

func NewOverallClient(apiURL string) *OverallClient {
	return &OverallClient{
		BaseAPIClient: NewBaseAPIClient(apiURL),
	}
}

func (c *OverallClient) Evaluate(ctx context.Context, essay map[string]any) (*dto_evaluate.APIOverall, error) {
	var response dto_evaluate.APIOverall
	if err := c.client.Post(ctx, c.apiURL, essay, &response); err != nil {
		return nil, fmt.Errorf("总体评价失败: %w", err)
	}
	return &response, nil
}

// ExpressionClient 表达评估客户端
type ExpressionClient struct {
	*BaseAPIClient
}

func NewExpressionClient(apiURL string) *ExpressionClient {
	return &ExpressionClient{
		BaseAPIClient: NewBaseAPIClient(apiURL),
	}
}

func (c *ExpressionClient) Evaluate(ctx context.Context, essay map[string]any) (*dto_evaluate.APIExpression, error) {
	var response dto_evaluate.APIExpression
	if err := c.client.Post(ctx, c.apiURL, essay, &response); err != nil {
		return nil, fmt.Errorf("表达评估失败: %w", err)
	}
	return &response, nil
}

// SuggestionClient 建议生成客户端
type SuggestionClient struct {
	*BaseAPIClient
}

func NewSuggestionClient(apiURL string) *SuggestionClient {
	return &SuggestionClient{
		BaseAPIClient: NewBaseAPIClient(apiURL),
	}
}

func (c *SuggestionClient) Generate(ctx context.Context, essay map[string]any) (*dto_evaluate.APISuggestion, error) {
	var response dto_evaluate.APISuggestion
	if err := c.client.Post(ctx, c.apiURL, essay, &response); err != nil {
		return nil, fmt.Errorf("建议生成失败: %w", err)
	}
	return &response, nil
}

// ParagraphClient 段落评估客户端
type ParagraphClient struct {
	*BaseAPIClient
}

func NewParagraphClient(apiURL string) *ParagraphClient {
	return &ParagraphClient{
		BaseAPIClient: NewBaseAPIClient(apiURL),
	}
}

func (c *ParagraphClient) Evaluate(ctx context.Context, essay map[string]any) (*dto_evaluate.APIParagraph, error) {
	var response dto_evaluate.APIParagraph
	if err := c.client.Post(ctx, c.apiURL, essay, &response); err != nil {
		return nil, fmt.Errorf("段落评估失败: %w", err)
	}
	return &response, nil
}

// ScoreClient 评分客户端
type ScoreClient struct {
	*BaseAPIClient
}

func NewScoreClient(apiURL string) *ScoreClient {
	return &ScoreClient{
		BaseAPIClient: NewBaseAPIClient(apiURL),
	}
}

func (c *ScoreClient) Calculate(ctx context.Context, essay map[string]any, req *model.EvaluateRequest) (*model.APIScore, error) {
	scoreEssay := make(map[string]any)
	if err := copier.Copy(&scoreEssay, &essay); err != nil {
		return nil, err
	}

	// 设置可选字段
	if req.Prompt != nil {
		scoreEssay["prompt"] = *req.Prompt
	}
	if req.Standard != nil {
		scoreEssay["rubric"] = *req.Standard
	}

	// 构建自定义分项打分比例
	if req.ContentScore != nil || req.ExpressionScore != nil ||
		req.StructureScore != nil || req.DevelopmentScore != nil {
		ratio := make(map[string]any)
		if req.ContentScore != nil {
			ratio["content"] = *req.ContentScore
		}
		if req.ExpressionScore != nil {
			ratio["expression"] = *req.ExpressionScore
		}
		if req.StructureScore != nil {
			ratio["structure"] = *req.StructureScore
		}
		if req.DevelopmentScore != nil {
			ratio["development"] = *req.DevelopmentScore
		}
		scoreEssay["ratio"] = ratio
	}

	scoreEssay["image"] = ""
	scoreEssay["type"] = "essay"

	var response model.APIScore
	if err := c.client.Post(ctx, c.apiURL, scoreEssay, &response); err != nil {
		return nil, fmt.Errorf("评分计算失败: %w", err)
	}
	return &response, nil
}

// PolishingClient 润色客户端
type PolishingClient struct {
	*BaseAPIClient
}

func NewPolishingClient(apiURL string) *PolishingClient {
	return &PolishingClient{
		BaseAPIClient: NewBaseAPIClient(apiURL),
	}
}

func (c *PolishingClient) Polish(ctx context.Context, essay map[string]any) (*model.APIPolishingContent, error) {
	var response model.APIPolishingContent
	if err := c.client.Post(ctx, c.apiURL, essay, &response); err != nil {
		return nil, fmt.Errorf("内容润色失败: %w", err)
	}
	return &response, nil
}

// PolishStream 流式润色（返回channel接收流式数据）
func (c *PolishingClient) PolishStream(ctx context.Context, essay map[string]any, resultChan chan<- string) error {
	if err := c.client.PostWithStream(ctx, c.apiURL, nil, essay, resultChan); err != nil {
		return fmt.Errorf("流式润色失败: %w", err)
	}
	return nil
}
