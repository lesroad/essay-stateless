package model

import (
	"encoding/json"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func NewSuccessResponse(data any) *Response {
	return &Response{
		Code:    0,
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
	OverallEvaluation      OverallEvaluation      `json:"overallEvaluation"`      // 总评
	FluencyEvaluation      FluencyEvaluation      `json:"fluencyEvaluation"`      // 流畅度评价
	WordSentenceEvaluation WordSentenceEvaluation `json:"wordSentenceEvaluation"` // 好词好句评价
	ExpressionEvaluation   ExpressionEvaluation   `json:"expressionEvaluation"`   // 逻辑表达评价
	SuggestionEvaluation   SuggestionEvaluation   `json:"suggestionEvaluation"`   // 建议
	ParagraphEvaluations   []ParagraphEvaluation  `json:"paragraphEvaluations"`   // 段落点评
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

/*
	{
		"wordSentenceEvaluation": {
			"sentenceEvaluations": [
				[{
					"isGoodSentence": false,
					"label": "",
					"type": {},
					"wordEvaluations": [{
						"ori": "，",
						"revised": "、",
						"span": [55, 56],
						"type": {
							"level1": "还需努力",
							"level2": "标点问题"
						}
					}, {
						"ori": "，",
						"revised": "。",
						"span": [64, 65],
						"type": {
							"level1": "还需努力",
							"level2": "标点问题"
						}
					}]
				}],
				[{
					"isGoodSentence": false,
					"label": "",
					"type": {},
					"wordEvaluations": []
				}, {
					"isGoodSentence": true,
					"label": "排比",
					"type": {
						"level1": "作文亮点",
						"level2": "好句"
					},
					"wordEvaluations": [{
						"span": [7, 11],
						"type": {
							"level1": "作文亮点",
							"level2": "好词"
						}
					}]
				}, {
					"isGoodSentence": true,
					"label": "比拟",
					"type": {
						"level1": "作文亮点",
						"level2": "好句"
					},
					"wordEvaluations": []
				}, {
					"isGoodSentence": false,
					"label": "",
					"type": {},
					"wordEvaluations": []
				}, {
					"isGoodSentence": true,
					"label": "比拟",
					"type": {
						"level1": "作文亮点",
						"level2": "好句"
					},
					"wordEvaluations": []
				}, {
					"isGoodSentence": false,
					"label": "",
					"type": {},
					"wordEvaluations": []
				}]
			],
			"wordSentenceScore": 0
		}
	}
*/
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

/*
	"expressionEvaluation": {
				"expressDescription": "作文能围绕主题展开，但逻辑表达较松散，缺乏层次感。描述场景时重复信息较多（如“放风筝”），未能有效组织细节。人物活动描写琐碎，未形成连贯叙事。建议学习如何筛选关键细节，构建更有条理的场景描写，避免重复和碎片化表达。",
				"expressionScore": 2
			},
*/
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
