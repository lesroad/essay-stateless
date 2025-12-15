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
	AllScore  int64    `json:"score"`
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
	WordSentenceEvaluation WordSentenceEvaluation `json:"wordSentenceEvaluation,omitempty"` // 好词好句评价
	SuggestionEvaluation   SuggestionEvaluation   `json:"suggestionEvaluation,omitempty"`   // 建议
	ParagraphEvaluations   []ParagraphEvaluation  `json:"paragraphEvaluations,omitempty"`   // 段落点评
	ScoreEvaluation        ScoreEvaluation        `json:"scoreEvaluations,omitempty"`       // 分数点评
	PolishingEvaluation    []PolishingEvaluation  `json:"polishingEvaluation,omitempty"`    // 润色点评
}

type ModelVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type OverallEvaluation struct {
	Description         string `json:"description"`
	TopicRelevanceScore int    `json:"topicRelevanceScore"`
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
	All         int64 `json:"all"`
	Appearance  int64 `json:"appearance"`
	Content     int64 `json:"content"`
	Expression  int64 `json:"expression"`
	Structure   int64 `json:"structure,omitempty"`
	Development int64 `json:"development,omitempty"`
	// 分项分数 / 总分
	AllWithTotal         string `json:"allWithTotal"`
	ContentWithTotal     string `json:"contentWithTotal"`
	ExpressionWithTotal  string `json:"expressionWithTotal"`
	StructureWithTotal   string `json:"structureWithTotal"`
	DevelopmentWithTotal string `json:"developmentWithTotal"`
}

type PolishingEvaluation struct {
	ParagraphIndex int             `json:"paragraphIndex"`
	Edits          []PolishingEdit `json:"edits"`
}

type PolishingEdit struct {
	Op            string `json:"op"`
	Reason        string `json:"reason"`
	Original      string `json:"original"`
	Revised       string `json:"revised,omitempty"`
	SentenceIndex int    `json:"sentenceIndex"`
	Span          []int  `json:"span"`
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
	Type      string `json:"type"`      // 响应类型: "init", "progress", "complete", "error"
	Step      string `json:"step"`      // 当前步骤
	Progress  int    `json:"progress"`  // 进度百分比 (0-100)
	Data      any    `json:"data"`      // 具体数据
	Message   string `json:"message"`   // 状态消息
	Timestamp int64  `json:"timestamp"` // 时间戳
}

// StreamInitData 初始化数据
type StreamInitData struct {
	Title     string     `json:"title"`
	Text      [][]string `json:"text"`
	EssayInfo EssayInfo  `json:"essay_info"` // 这里要改下todo
}

// StreamStepData 步骤完成数据
type StreamStepData struct {
	Step string `json:"step"`
	Data any    `json:"data"`
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

// ClassStatisticsResponse 班级学情统计分析响应
type ClassStatisticsResponse struct {
	TotalStudents      int                `json:"totalStudents"`      // 总学生数
	OverallPerformance OverallPerformance `json:"overallPerformance"` // 整体表现
	ErrorAnalysis      ErrorAnalysis      `json:"errorAnalysis"`      // 错误分析
	HighlightAnalysis  HighlightAnalysis  `json:"highlightAnalysis"`  // 亮点分析
	GeneratedTime      int64              `json:"generatedTime"`      // 生成时间戳
}

// OverallPerformance 整体表现分析
type OverallPerformance struct {
	AverageScore         float64                 `json:"averageScore"`         // 平均分
	GradeDistribution    []GradeDistributionItem `json:"gradeDistribution"`    // 等级分布：优秀、良好、合格、不合格分别多少人
	SkillMasteryAnalysis []SkillMasteryItem      `json:"skillMasteryAnalysis"` // 写作技能掌握情况：各项技能在不同等级中的分布
	Summary              string                  `json:"summary"`              // 整体表现总结
}

// GradeDistributionItem 等级分布项
type GradeDistributionItem struct {
	Grade        string  `json:"grade"`        // 等级名称：优秀、良好、合格、不合格
	StudentCount int     `json:"studentCount"` // 该等级的学生人数
	Percentage   float64 `json:"percentage"`   // 占总学生数的百分比
}

// SkillMasteryItem 写作技能掌握情况项
type SkillMasteryItem struct {
	SkillName         string                   `json:"skillName"`         // 技能名称：All、Appearance、Content、Expression、Structure、Development
	GradeDistribution []SkillGradeDistribution `json:"gradeDistribution"` // 该技能在各等级中的分布
}

// SkillGradeDistribution 技能等级分布
type SkillGradeDistribution struct {
	Grade        string  `json:"grade"`        // 等级名称
	StudentCount int     `json:"studentCount"` // 该等级的学生人数
	Percentage   float64 `json:"percentage"`   // 占总学生数的百分比
}

// ErrorAnalysis 错误分析
type ErrorAnalysis struct {
	ErrorDistribution []ErrorDistributionItem `json:"errorDistribution"` // 错误个数分布：0个到6个及以上分别有多少人
	ErrorTypeRatio    []ErrorTypeItem         `json:"errorTypeRatio"`    // 错误类型占比：每种错误类型占比多少，分别是多少人次
	HighFrequencyList []HighFrequencyError    `json:"highFrequencyList"` // 高频错误列表：哪些错误分别有多少人次犯了，从多到少排序
}

// ErrorDistributionItem 错误个数分布项
type ErrorDistributionItem struct {
	ErrorCount   string  `json:"errorCount"`   // 错误个数范围，如"0个", "1个", "2个", ..., "6个及以上"
	StudentCount int     `json:"studentCount"` // 该错误个数范围的学生人数
	Percentage   float64 `json:"percentage"`   // 占总学生数的百分比
}

// ErrorTypeItem 错误类型占比项
type ErrorTypeItem struct {
	ErrorType    string  `json:"errorType"`    // 错误类型名称
	Count        int     `json:"count"`        // 该类型错误总人次
	Percentage   float64 `json:"percentage"`   // 占所有错误的百分比
	StudentCount int     `json:"studentCount"` // 犯该类型错误的学生人数
}

// HighFrequencyError 高频错误项
type HighFrequencyError struct {
	ErrorText string   `json:"errorText"` // 具体错误内容
	ErrorType string   `json:"errorType"` // 错误类型
	Count     int      `json:"count"`     // 犯该错误的人次
	Examples  []string `json:"examples"`  // 错误示例（原文->修改后）
}

// HighlightAnalysis 亮点分析
type HighlightAnalysis struct {
	HighlightDistribution []HighlightDistributionItem `json:"highlightDistribution"` // 亮点个数分布：0个到6个及以上分别有多少人
	HighlightTypeRatio    []HighlightTypeItem         `json:"highlightTypeRatio"`    // 亮点类型占比：每种亮点类型占比多少，分别是多少人次
}

// HighlightDistributionItem 亮点个数分布项
type HighlightDistributionItem struct {
	HighlightCount string  `json:"highlightCount"` // 亮点个数范围，如"0个", "1个", "2个", ..., "6个及以上"
	StudentCount   int     `json:"studentCount"`   // 该亮点个数范围的学生人数
	Percentage     float64 `json:"percentage"`     // 占总学生数的百分比
}

// HighlightTypeItem 亮点类型占比项
type HighlightTypeItem struct {
	HighlightType string  `json:"highlightType"` // 亮点类型名称
	Count         int     `json:"count"`         // 该类型亮点总人次
	Percentage    float64 `json:"percentage"`    // 占所有亮点的百分比
	StudentCount  int     `json:"studentCount"`  // 有该类型亮点的学生人数
}

func (r *ClassStatisticsResponse) JSONString() string {
	data, _ := json.Marshal(r)
	return string(data)
}
