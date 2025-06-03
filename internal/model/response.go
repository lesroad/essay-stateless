package model

import (
	"encoding/json"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewSuccessResponse(data interface{}) *Response {
	return &Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

func NewErrorResponse(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

type BetaEvaluateResponse struct {
	Title        string       `json:"title"`
	Text         [][]string   `json:"text"`
	EssayInfo    EssayInfo    `json:"essayInfo"`
	AIEvaluation AIEvaluation `json:"aiEvaluation"`
}

func (r *BetaEvaluateResponse) JSONString() string {
	data, _ := json.Marshal(r)
	return string(data)
}

type EssayInfo struct {
	EssayType string   `json:"essayType"`
	Grade     int      `json:"grade"`
	Counting  Counting `json:"counting"`
}

type Counting struct {
	AdjAdvNum         int `json:"adjAdvNum"`
	CharNum           int `json:"charNum"`
	DieciNum          int `json:"dieciNum"`
	Fluency           int `json:"fluency"`
	GrammarMistakeNum int `json:"grammarMistakeNum"`
	HighlightSentsNum int `json:"highlightSentsNum"`
	IdiomNum          int `json:"idiomNum"`
	NounTypeNum       int `json:"nounTypeNum"`
	ParaNum           int `json:"paraNum"`
	SentNum           int `json:"sentNum"`
	UniqueWordNum     int `json:"uniqueWordNum"`
	VerbTypeNum       int `json:"verbTypeNum"`
	WordNum           int `json:"wordNum"`
	WrittenMistakeNum int `json:"writtenMistakeNum"`
}

type AIEvaluation struct {
	ModelVersion           ModelVersion           `json:"modelVersion"`
	OverallEvaluation      OverallEvaluation      `json:"overallEvaluation"`
	FluencyEvaluation      FluencyEvaluation      `json:"fluencyEvaluation"`
	WordSentenceEvaluation WordSentenceEvaluation `json:"wordSentenceEvaluation"`
	ExpressionEvaluation   ExpressionEvaluation   `json:"expressionEvaluation"`
	SuggestionEvaluation   SuggestionEvaluation   `json:"suggestionEvaluation"`
	ParagraphEvaluations   []ParagraphEvaluation  `json:"paragraphEvaluations"`
}

type ModelVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type OverallEvaluation struct {
	Description         string `json:"description"`
	TopicRelevanceScore int    `json:"topicRelevanceScore"`
}

type FluencyEvaluation struct {
	FluencyDescription string `json:"fluencyDescription"`
	FluencyScore       int    `json:"fluencyScore"`
}

type WordSentenceEvaluation struct {
	SentenceEvaluations [][]SentenceEvaluation `json:"sentenceEvaluations"`
	WordSentenceScore   int                    `json:"wordSentenceScore"`
}

type SentenceEvaluation struct {
	IsGoodSentence  bool              `json:"isGoodSentence"`
	Label           string            `json:"label"`
	Type            map[string]string `json:"type"`
	WordEvaluations []WordEvaluation  `json:"wordEvaluations"`
}

type WordEvaluation struct {
	Span    []int             `json:"span"`
	Type    map[string]string `json:"type"`
	Ori     string            `json:"ori,omitempty"`
	Revised string            `json:"revised,omitempty"`
}

type ExpressionEvaluation struct {
	ExpressDescription string `json:"expressDescription"`
	ExpressionScore    int    `json:"expressionScore"`
}

type SuggestionEvaluation struct {
	SuggestionDescription string `json:"suggestionDescription"`
}

type ParagraphEvaluation struct {
	ParagraphIndex int    `json:"paragraphIndex"`
	Comment        string `json:"comment"`
}

type TitleOcrResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (r *TitleOcrResponse) JSONString() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type DefaultOcrResponse struct {
	Title   *string `json:"title"`
	Content string  `json:"content"`
}

func (r *DefaultOcrResponse) JSONString() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
