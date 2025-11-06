package statistics

// StudentData 学生数据（简化的内部表示）
type StudentData struct {
	WordSentenceEvaluation WordSentenceEvaluation
	Scores                 ScoreData
}

// WordSentenceEvaluation 好词好句评价
type WordSentenceEvaluation struct {
	SentenceEvaluations [][]SentenceEvaluation `json:"sentenceEvaluations"`
}

// SentenceEvaluation 句子评价
type SentenceEvaluation struct {
	IsGoodSentence  bool              `json:"isGoodSentence"`
	Label           string            `json:"label"`
	Type            map[string]string `json:"type"`
	WordEvaluations []WordEvaluation  `json:"wordEvaluations"`
}

// WordEvaluation 词汇评价
type WordEvaluation struct {
	Span    []int             `json:"span"`
	Type    map[string]string `json:"type"`
	Ori     string            `json:"ori"`
	Revised string            `json:"revised"`
}

// ScoreData 分数数据
type ScoreData struct {
	All         int
	Content     int
	Expression  int
	Structure   int
	Development int
}

// OverallPerformance 整体表现
type OverallPerformance struct {
	AverageScore         float64                 `json:"averageScore"`
	GradeDistribution    []GradeDistributionItem `json:"gradeDistribution"`
	SkillMasteryAnalysis []SkillMasteryItem      `json:"skillMasteryAnalysis"`
	Summary              string                  `json:"summary"`
}

// GradeDistributionItem 等级分布项
type GradeDistributionItem struct {
	Grade        string  `json:"grade"`
	StudentCount int     `json:"studentCount"`
	Percentage   float64 `json:"percentage"`
}

// SkillMasteryItem 技能掌握情况项
type SkillMasteryItem struct {
	SkillName         string                   `json:"skillName"`
	GradeDistribution []SkillGradeDistribution `json:"gradeDistribution"`
}

// SkillGradeDistribution 技能等级分布
type SkillGradeDistribution struct {
	Grade        string  `json:"grade"`
	StudentCount int     `json:"studentCount"`
	Percentage   float64 `json:"percentage"`
}

// ErrorAnalysis 错误分析
type ErrorAnalysis struct {
	ErrorDistribution []ErrorDistributionItem `json:"errorDistribution"`
	ErrorTypeRatio    []ErrorTypeItem         `json:"errorTypeRatio"`
	HighFrequencyList []HighFrequencyError    `json:"highFrequencyList"`
}

// ErrorDistributionItem 错误个数分布项
type ErrorDistributionItem struct {
	ErrorCount   string  `json:"errorCount"`
	StudentCount int     `json:"studentCount"`
	Percentage   float64 `json:"percentage"`
}

// ErrorTypeItem 错误类型占比项
type ErrorTypeItem struct {
	ErrorType    string  `json:"errorType"`
	Count        int     `json:"count"`
	Percentage   float64 `json:"percentage"`
	StudentCount int     `json:"studentCount"`
}

// HighFrequencyError 高频错误项
type HighFrequencyError struct {
	ErrorText string   `json:"errorText"`
	ErrorType string   `json:"errorType"`
	Count     int      `json:"count"`
	Examples  []string `json:"examples"`
}

// HighlightAnalysis 亮点分析
type HighlightAnalysis struct {
	HighlightDistribution []HighlightDistributionItem `json:"highlightDistribution"`
	HighlightTypeRatio    []HighlightTypeItem         `json:"highlightTypeRatio"`
}

// HighlightDistributionItem 亮点个数分布项
type HighlightDistributionItem struct {
	HighlightCount string  `json:"highlightCount"`
	StudentCount   int     `json:"studentCount"`
	Percentage     float64 `json:"percentage"`
}

// HighlightTypeItem 亮点类型占比项
type HighlightTypeItem struct {
	HighlightType string  `json:"highlightType"`
	Count         int     `json:"count"`
	Percentage    float64 `json:"percentage"`
	StudentCount  int     `json:"studentCount"`
}

