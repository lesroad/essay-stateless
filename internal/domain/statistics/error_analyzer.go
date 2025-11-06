package statistics

import (
	"essay-stateless/internal/dto/statistics"
	"sort"
)

// ErrorAnalyzer 错误分析器
type ErrorAnalyzer struct {
	gradeCalc *GradeCalculator
}

// NewErrorAnalyzer 创建错误分析器
func NewErrorAnalyzer(gradeCalc *GradeCalculator) *ErrorAnalyzer {
	return &ErrorAnalyzer{
		gradeCalc: gradeCalc,
	}
}

// Analyze 分析错误情况
func (a *ErrorAnalyzer) Analyze(students []statistics.StudentData) statistics.ErrorAnalysis {
	totalStudents := len(students)
	if totalStudents == 0 {
		return statistics.ErrorAnalysis{}
	}

	// 统计数据
	studentErrorCounts := make([]int, totalStudents)
	errorTypeCount := make(map[string]int)
	errorTypeStudents := make(map[string]map[int]bool)
	specificErrorCount := make(map[string]int)
	specificErrorType := make(map[string]string)
	specificErrorExamples := make(map[string][]string)

	// 遍历学生数据
	for idx, student := range students {
		errorCount := 0
		errorTypes := make(map[string]bool)

		for _, paragraphs := range student.WordSentenceEvaluation.SentenceEvaluations {
			for _, sentence := range paragraphs {
				for _, wordEval := range sentence.WordEvaluations {
					if level1, ok := wordEval.Type["level1"]; ok && level1 == "还需努力" {
						if level2, ok := wordEval.Type["level2"]; ok {
							errorCount++
							errorTypeCount[level2]++

							// 记录学生犯过的错误类型
							if !errorTypes[level2] {
								errorTypes[level2] = true
								if errorTypeStudents[level2] == nil {
									errorTypeStudents[level2] = make(map[int]bool)
								}
								errorTypeStudents[level2][idx] = true
							}

							// 统计具体错误
							if wordEval.Ori != "" {
								specificErrorCount[wordEval.Ori]++
								specificErrorType[wordEval.Ori] = level2

								examples := specificErrorExamples[wordEval.Ori]
								if len(examples) < 3 {
									example := wordEval.Ori
									if wordEval.Revised != "" {
										example = wordEval.Ori + " -> " + wordEval.Revised
									}
									specificErrorExamples[wordEval.Ori] = append(examples, example)
								}
							}
						}
					}
				}
			}
		}

		studentErrorCounts[idx] = errorCount
	}

	return statistics.ErrorAnalysis{
		ErrorDistribution: a.generateErrorDistribution(studentErrorCounts, totalStudents),
		ErrorTypeRatio:    a.generateErrorTypeRatio(errorTypeCount, errorTypeStudents, totalStudents),
		HighFrequencyList: a.generateHighFrequencyList(specificErrorCount, specificErrorType, specificErrorExamples),
	}
}

func (a *ErrorAnalyzer) generateErrorDistribution(counts []int, total int) []statistics.ErrorDistributionItem {
	distribution := make([]int, 7) // 0个, 1个, 2个, 3个, 4个, 5个, 6个及以上

	for _, count := range counts {
		if count >= 6 {
			distribution[6]++
		} else {
			distribution[count]++
		}
	}

	labels := []string{"0个", "1个", "2个", "3个", "4个", "5个", "6个及以上"}
	result := make([]statistics.ErrorDistributionItem, 0, len(labels))

	for i, count := range distribution {
		result = append(result, statistics.ErrorDistributionItem{
			ErrorCount:   labels[i],
			StudentCount: count,
			Percentage:   a.gradeCalc.RoundPercentage(float64(count) / float64(total) * 100),
		})
	}

	return result
}

func (a *ErrorAnalyzer) generateErrorTypeRatio(typeCount map[string]int, typeStudents map[string]map[int]bool, total int) []statistics.ErrorTypeItem {
	totalCount := 0
	for _, count := range typeCount {
		totalCount += count
	}

	if totalCount == 0 {
		return []statistics.ErrorTypeItem{}
	}

	result := make([]statistics.ErrorTypeItem, 0, len(typeCount))
	for errorType, count := range typeCount {
		result = append(result, statistics.ErrorTypeItem{
			ErrorType:    errorType,
			Count:        count,
			Percentage:   a.gradeCalc.RoundPercentage(float64(count) / float64(totalCount) * 100),
			StudentCount: len(typeStudents[errorType]),
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	return result
}

func (a *ErrorAnalyzer) generateHighFrequencyList(errorCount map[string]int, errorType map[string]string, examples map[string][]string) []statistics.HighFrequencyError {
	result := make([]statistics.HighFrequencyError, 0, len(errorCount))

	for text, count := range errorCount {
		result = append(result, statistics.HighFrequencyError{
			ErrorText: text,
			ErrorType: errorType[text],
			Count:     count,
			Examples:  examples[text],
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	if len(result) > 20 {
		result = result[:20]
	}

	return result
}

