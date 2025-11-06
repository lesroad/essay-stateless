package statistics

import "essay-stateless/internal/dto/statistics"

// Analyzer 统计分析器（主入口）
type Analyzer struct {
	gradeCalc         *GradeCalculator
	overallAnalyzer   *OverallAnalyzer
	errorAnalyzer     *ErrorAnalyzer
	highlightAnalyzer *HighlightAnalyzer
}

// NewAnalyzer 创建统计分析器
func NewAnalyzer() *Analyzer {
	gradeCalc := NewGradeCalculator()

	return &Analyzer{
		gradeCalc:         gradeCalc,
		overallAnalyzer:   NewOverallAnalyzer(gradeCalc),
		errorAnalyzer:     NewErrorAnalyzer(gradeCalc),
		highlightAnalyzer: NewHighlightAnalyzer(gradeCalc),
	}
}

// AnalyzeResult 分析结果
type AnalyzeResult struct {
	TotalStudents      int                           `json:"totalStudents"`
	OverallPerformance statistics.OverallPerformance `json:"overallPerformance"`
	ErrorAnalysis      statistics.ErrorAnalysis      `json:"errorAnalysis"`
	HighlightAnalysis  statistics.HighlightAnalysis  `json:"highlightAnalysis"`
}

// Analyze 执行完整的学情分析
func (a *Analyzer) Analyze(students []statistics.StudentData) *AnalyzeResult {
	if len(students) == 0 {
		return &AnalyzeResult{}
	}

	return &AnalyzeResult{
		TotalStudents:      len(students),
		OverallPerformance: a.overallAnalyzer.Analyze(students),
		ErrorAnalysis:      a.errorAnalyzer.Analyze(students),
		HighlightAnalysis:  a.highlightAnalyzer.Analyze(students),
	}
}

