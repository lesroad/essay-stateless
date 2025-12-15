package evaluate

import "essay-stateless/internal/config"

// APIClientsFactory API客户端工厂
type APIClientsFactory struct {
	apiConfig *config.EvaluateAPIConfig
}

// NewAPIClientsFactory 创建API客户端工厂
func NewAPIClientsFactory(apiConfig *config.EvaluateAPIConfig) *APIClientsFactory {
	return &APIClientsFactory{
		apiConfig: apiConfig,
	}
}

// CreateEssayInfoClient 创建作文信息客户端
func (f *APIClientsFactory) CreateEssayInfoClient() *EssayInfoClient {
	return NewEssayInfoClient(f.apiConfig.EssayInfo)
}

// CreateWordSentenceClient 创建词句评估客户端
func (f *APIClientsFactory) CreateWordSentenceClient() *WordSentenceClient {
	return NewWordSentenceClient(f.apiConfig.WordSentence)
}

// CreateGrammarClient 创建语法检查客户端
func (f *APIClientsFactory) CreateGrammarClient() *GrammarClient {
	return NewGrammarClient(f.apiConfig.GrammarInfo)
}

// CreateOverallClient 创建总体评价客户端
func (f *APIClientsFactory) CreateOverallClient() *OverallClient {
	return NewOverallClient(f.apiConfig.Overall)
}

// CreateSuggestionClient 创建建议生成客户端
func (f *APIClientsFactory) CreateSuggestionClient() *SuggestionClient {
	return NewSuggestionClient(f.apiConfig.Suggestion)
}

// CreateParagraphClient 创建段落评估客户端
func (f *APIClientsFactory) CreateParagraphClient() *ParagraphClient {
	return NewParagraphClient(f.apiConfig.Paragraph)
}

// CreateScoreClient 创建评分客户端
func (f *APIClientsFactory) CreateScoreClient() *ScoreClient {
	return NewScoreClient(f.apiConfig.Score)
}

// CreatePolishingClient 创建润色客户端
func (f *APIClientsFactory) CreatePolishingClient() *PolishingClient {
	return NewPolishingClient(f.apiConfig.Polishing)
}
