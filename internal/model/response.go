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

type EvaluateResponse struct {
	Title        string       `json:"title"`
	Text         [][]string   `json:"text"`
	EssayInfo    EssayInfo    `json:"essayInfo"`
	AIEvaluation AIEvaluation `json:"aiEvaluation"`
}

func (r *EvaluateResponse) JSONString() string {
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
	ModelVersion           ModelVersion           `json:"modelVersion,omitempty"`
	OverallEvaluation      OverallEvaluation      `json:"overallEvaluation,omitempty"`      // 总评
	FluencyEvaluation      FluencyEvaluation      `json:"fluencyEvaluation,omitempty"`      // 流畅度评价
	WordSentenceEvaluation WordSentenceEvaluation `json:"wordSentenceEvaluation,omitempty"` // 好词好句评价
	ExpressionEvaluation   ExpressionEvaluation   `json:"expressionEvaluation,omitempty"`   // 逻辑表达评价
	SuggestionEvaluation   SuggestionEvaluation   `json:"suggestionEvaluation,omitempty"`   // 建议
	ParagraphEvaluations   []ParagraphEvaluation  `json:"paragraphEvaluations,omitempty"`   // 段落点评
	ScoreEvaluation        ScoreEvaluation        `json:"scoreEvaluations,omitempty"`       // 分数点评
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
	Type            map[string]string `json:"type"`            // 好句类型
	WordEvaluations []WordEvaluation  `json:"wordEvaluations"` // 好词/还需努力的词
}

type WordEvaluation struct {
	Span    []int             `json:"span"`
	Type    map[string]string `json:"type"`
	Ori     string            `json:"ori,omitempty"`
	Revised string            `json:"revised,omitempty"`
}

/*
	"expressionEvaluation": {
				"expressDescription": "作文能围绕主题展开，但逻辑表达较松散，缺乏层次感。描述场景时重复信息较多（如"放风筝"），未能有效组织细节。人物活动描写琐碎，未形成连贯叙事。建议学习如何筛选关键细节，构建更有条理的场景描写，避免重复和碎片化表达。",
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

type ScoreEvaluation struct {
	Comment  string   `json:"comment"`
	Comments Comments `json:"comments"`
	Scores   Scores   `json:"scores"`
}

type Comments struct {
	Appearance  string `json:"appearance"`
	Content     string `json:"content"`
	Expression  string `json:"expression"`
	Structure   string `json:"structure,omitempty"`   // 结构-初中
	Development string `json:"development,omitempty"` // 发展-高中
}

type Scores struct {
	All         int `json:"all"`
	Appearance  int `json:"appearance"`
	Content     int `json:"content"`
	Expression  int `json:"expression"`
	Structure   int `json:"structure,omitempty"`
	Development int `json:"development,omitempty"`
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

// StreamEvaluateResponse 流式评估响应
type StreamEvaluateResponse struct {
	Type      string      `json:"type"`      // 响应类型: "init", "progress", "complete", "error"
	Step      string      `json:"step"`      // 当前步骤: "essay_info", "overall", "fluency", "word_sentence", "expression", "suggestion", "paragraph", "grammar"
	Progress  int         `json:"progress"`  // 进度百分比 (0-100)
	Data      interface{} `json:"data"`      // 具体数据
	Message   string      `json:"message"`   // 状态消息
	Timestamp int64       `json:"timestamp"` // 时间戳
}

// StreamInitData 初始化数据
type StreamInitData struct {
	Title     string     `json:"title"`
	Text      [][]string `json:"text"`
	EssayInfo EssayInfo  `json:"essay_info"`
}

// StreamStepData 步骤完成数据
type StreamStepData struct {
	Step string      `json:"step"`
	Data interface{} `json:"data"`
}

// StreamCompleteData 完成数据
type StreamCompleteData struct {
	Result *EvaluateResponse `json:"result"`
}

// StreamErrorData 错误数据
type StreamErrorData struct {
	Error string `json:"error"`
	Step  string `json:"step"`
}

// JSONString 序列化流式响应
func (r *StreamEvaluateResponse) JSONString() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
