package statistics

import "essay-stateless/internal/dto/statistics"

type Analyzer struct {
	gradeCalc         *GradeCalculator
	overallAnalyzer   *OverallAnalyzer
	errorAnalyzer     *ErrorAnalyzer
	highlightAnalyzer *HighlightAnalyzer
}

func NewAnalyzer() *Analyzer {
	gradeCalc := NewGradeCalculator()

	return &Analyzer{
		gradeCalc:         gradeCalc,
		overallAnalyzer:   NewOverallAnalyzer(gradeCalc),
		errorAnalyzer:     NewErrorAnalyzer(gradeCalc),
		highlightAnalyzer: NewHighlightAnalyzer(gradeCalc),
	}
}

type AnalyzeResult struct {
	SubmissionPercentage float64                       `json:"submissionPercentage"`
	OverallPerformance   statistics.OverallPerformance `json:"overallPerformance"`
	ErrorAnalysis        statistics.ErrorAnalysis      `json:"errorAnalysis"`
	HighlightAnalysis    statistics.HighlightAnalysis  `json:"highlightAnalysis"`
}

func (a *Analyzer) Analyze(students []statistics.StudentData, totalStudents int) *AnalyzeResult {
	return &AnalyzeResult{
		SubmissionPercentage: float64(len(students)) / float64(totalStudents) * 100,
		OverallPerformance:   a.overallAnalyzer.Analyze(students),
		ErrorAnalysis:        a.errorAnalyzer.Analyze(students),
		HighlightAnalysis:    a.highlightAnalyzer.Analyze(students),
	}
}
